/*
* Copyright (c) 2023, Intel Corporation. All rights reserved.<BR>
* SPDX-License-Identifier: Apache-2.0
*/

use anyhow::*;
use clap::Parser;
use core::result::Result::Ok;
use hyper::{Request as HyperRequest, Response as HyperResponse, Body, Server as HyperServer};
use hyper::service::{make_service_fn, service_fn};
use quote_server::get_quote_server::{GetQuote, GetQuoteServer};
use quote_server::{GetQuoteRequest, GetQuoteResponse};
use tokio::net::UnixListener;
use tokio_stream::wrappers::UnixListenerStream;
use tonic::{transport::Server, Request, Response, Status};

use std::fs::OpenOptions;
use std::io::Write;
use std::net::SocketAddr;
use std::os::unix::fs::OpenOptionsExt;

pub mod tee;
pub mod kube;
use tee::*;

pub mod quote_server {
    tonic::include_proto!("quoteserver");

    pub(crate) const FILE_DESCRIPTOR_SET: &[u8] =
        tonic::include_file_descriptor_set!("quote_server_descriptor");
}

pub struct CCNPGetQuote {
    local_tee: tee::TeeType,
}

impl CCNPGetQuote {
    fn new(_local_tee: TeeType) -> Self {
        CCNPGetQuote {
            local_tee: _local_tee,
        }
    }

    // generate current pod quote based on its all containers' imageIDs
    async fn gen_quote_by_current_pod_info(&self) {
        // Create an instance of your custom kube client
        let pod_data = kube::get_cur_pod_images_info();

        let _data = match pod_data.await {
            Ok(report_data) => {
                let report_data_clone = report_data.clone();
                let hash_report_data = kube::sha256_hash(&report_data_clone);
                let result = get_quote(self.local_tee.clone(), hash_report_data.clone(), hash_report_data.clone()).unwrap();

                println!("Current Pod Quote Data is below:");
                println!("--------------------------------");
                println!("{}", result);
                println!("--------------------------------");

                // Store the quote information in file
                let quote_file_path = "/run/ccnp/uds/quote.data";
                 // Create an OpenOptions object to configure the file creation
                let mut options = OpenOptions::new();

                // Set the desired permissions and other options
                options.write(true).create(true).mode(0o644);

                let mut quote_file = options.open(quote_file_path).unwrap();
                quote_file.write_all(result.as_bytes()).unwrap();
                // drop the quote file
                drop(quote_file);
            },
            Err(error) => {
                panic!("There was a problem when get current pod images information: {:?}", error)
            },
        };  
    }
}

#[tonic::async_trait]
impl GetQuote for CCNPGetQuote {
    async fn get_quote(
        &self,
        request: Request<GetQuoteRequest>,
    ) -> Result<Response<GetQuoteResponse>, Status> {
        let msg;
        let req = request.into_inner();

        println!(
            "Got a request with: user_data = {:?}, nonce = {:?}",
            req.user_data, req.nonce
        );

        self.gen_quote_by_current_pod_info().await;
        let result = get_quote(self.local_tee.clone(), req.user_data, req.nonce);
        match result {
            Ok(q) => {
                msg = Response::new(quote_server::GetQuoteResponse {
                    quote: q,
                    quote_type: format!("{:?}", self.local_tee).to_string(),
                })
            }
            Err(e) => return Err(Status::internal(e.to_string())),
        }
        Ok(msg)
    }
}

// A http server for provide the current pod quote data
#[derive(Copy, Clone)]
pub struct MyPodQuote {
    sock_address: SocketAddr,
    local_tee: tee::TeeType,
}

impl MyPodQuote {
    pub fn new(sock_address: SocketAddr, local_tee: tee::TeeType) -> Self {
        MyPodQuote {
            sock_address,
            local_tee,
        }
    }

    pub async fn start(&self) -> Result<(), hyper::Error> {
        let local_tee = self.local_tee;
        let make_svc = make_service_fn(|_conn| {
            let service = service_fn(move |req| {
                // Route request to the appropriate handler
                Self::handle_request(local_tee, req)
            });
            async move {
                Ok::<_, hyper::Error>(service)
            }
        });
        let http_server = HyperServer::bind(&self.sock_address).serve(make_svc);
        println!("The Pod Quote HTTP server is listening on: {:?}", self.sock_address);
        http_server.await
    }

    // generate current pod quote based on its all containers' imageIDs
    async fn get_current_pod_quote(local_tee: tee::TeeType) -> Result<String, anyhow::Error> {
        // Handle the "/quote" route
        // Create an instance of your custom kube client
        let pod_data = kube::get_cur_pod_images_info();
        match pod_data.await {
            Ok(report_data) => {
                let report_data_clone = report_data.clone();
                let hash_report_data = kube::sha256_hash(&report_data_clone);
                let quote_data = get_quote(local_tee, hash_report_data.clone(), hash_report_data.clone()).unwrap();
                Ok(quote_data)
            },
            Err(error) => {
                Err(anyhow!("There was a problem when get current pod images information: {:?}", error))
            },
        }
    }

    async fn handle_request(local_tee: tee::TeeType, req: HyperRequest<Body>) -> Result<HyperResponse<Body>, hyper::Error> {
        match req.uri().path() {
            "/quote" => {
                match Self::get_current_pod_quote(local_tee).await {
                    Ok(quote_data) => {
                        println!("File content: {}", quote_data);
                        // generate the response from quote file
                        let response = HyperResponse::new(Body::from(quote_data));
                        Ok(response)
                    },
                    Err(err) => {
                        eprintln!("Error: {}", err);
                        let response = HyperResponse::builder()
                            .status(404)
                            .body(Body::from("Not Found Quote File"))
                            .unwrap();
                        Ok(response)
                    },
                }
            }
            _ => {
                // Handle other routes
                let response = HyperResponse::builder()
                    .status(404)
                    .body(Body::from("Not Found"))
                    .unwrap();
                Ok(response)
            }
        }
    }
}

#[derive(Parser)]
struct Cli {
    port: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let uds_path = "/run/ccnp/uds/quote-server.sock";
    let http_addr = SocketAddr::from(([127, 0, 0, 1], 3000));

    let _ = std::fs::remove_file(uds_path);
    let uds = UnixListener::bind(uds_path)?;
    let uds_stream = UnixListenerStream::new(uds);

    let getquote = CCNPGetQuote::new({
        match tee::get_tee_type() {
            tee::TeeType::PLAIN => panic!("Not found any TEE device!"),
            t => t,
        }
    });

    let (mut health_reporter, health_service) = tonic_health::server::health_reporter();
    health_reporter
        .set_serving::<GetQuoteServer<CCNPGetQuote>>()
        .await;

    let reflection_service = tonic_reflection::server::Builder::configure()
        .register_encoded_file_descriptor_set(quote_server::FILE_DESCRIPTOR_SET)
        .build()
        .unwrap();

    println!(
        "Starting quote server in {} enviroment...",
        format!("{:?}", tee::get_tee_type()).to_string()
    );

    // Start the gRPC server
    let grpc_server_task = tokio::spawn(async move {
        let _grpc_server = Server::builder()
            .add_service(reflection_service)
            .add_service(health_service)
            .add_service(GetQuoteServer::new(getquote))
            .serve_with_incoming(uds_stream)
            .await
            .map_err(|e| eprintln!("gRPC server error: {}", e));
    });

    // Start http server for fetching quote with current pod image IDs
    let http_server_task = tokio::spawn(async move {
        let http_server = MyPodQuote::new(http_addr, 
            {
                match tee::get_tee_type() {
                    tee::TeeType::PLAIN => panic!("Not found any TEE device!"),
                    t => t,
                }
            }
        );
        if let Err(err) = http_server.start().await {
            eprintln!("HTTP server error: {}", err);
        }
    });

    // Wait for both grpc and http server tasks to finish
    if let (Err(grpc_error), Err(http_error)) = tokio::join!(grpc_server_task, http_server_task) {
        // You can handle the errors as needed, e.g., return an error result or log them.
        eprintln!("Error in gRPC server: {}", grpc_error);
        eprintln!("Error in HTTP server: {}", http_error);
        return Err("Error in one of the server tasks".into());
    }

    Ok(())
}

#[cfg(test)]
mod quote_server_tests {
    use super::*;
    use crate::quote_server::get_quote_client::GetQuoteClient;
    use serial_test::serial;
    use tokio::net::UnixStream;
    use tonic::transport::{Endpoint, Uri};
    use tower::service_fn;

    async fn creat_server() {
        let path = "/tmp/quote-server.sock";
        let _ = std::fs::remove_file(path);
        let uds = UnixListener::bind(path).unwrap();
        let uds_stream = UnixListenerStream::new(uds);

        let getquote = CCNPGetQuote::new({
            match tee::get_tee_type() {
                tee::TeeType::PLAIN => panic!("Not found any TEE device!"),
                t => t,
            }
        });

        tokio::spawn(async {
            Server::builder()
                .add_service(GetQuoteServer::new(getquote))
                .serve_with_incoming(uds_stream)
                .await
                .unwrap();
        });
    }

    #[tokio::test]
    #[serial]
    //test start server and send request
    async fn request_to_server_normal() {
        creat_server().await;

        let channel = Endpoint::try_from("http://[::]:40081")
            .unwrap()
            .connect_with_connector(service_fn(|_: Uri| {
                let path = "/tmp/quote-server.sock";
                UnixStream::connect(path)
            }))
            .await
            .unwrap();

        let mut client = GetQuoteClient::new(channel);

        let request = tonic::Request::new(GetQuoteRequest {
            user_data: base64::encode("123456781234567812345678123456781234567812345678"),
            nonce: "12345678".to_string(),
        });

        let response = client.get_quote(request).await.unwrap().into_inner();
        assert_eq!(response.quote_type, "TDX");
    }

    #[tokio::test]
    #[serial]
    async fn request_to_server_empty_user_data() {
        creat_server().await;

        let channel = Endpoint::try_from("http://[::]:40081")
            .unwrap()
            .connect_with_connector(service_fn(|_: Uri| {
                let path = "/tmp/quote-server.sock";
                UnixStream::connect(path)
            }))
            .await
            .unwrap();

        let mut client = GetQuoteClient::new(channel);

        let request = tonic::Request::new(GetQuoteRequest {
            user_data: "".to_string(),
            nonce: "12345678".to_string(),
        });

        let response = client.get_quote(request).await.unwrap().into_inner();
        assert_eq!(response.quote_type, "TDX");
        assert_ne!(response.quote.len(), 0);
    }

    #[tokio::test]
    #[serial]
    async fn request_to_server_long_user_data() {
        creat_server().await;

        let channel = Endpoint::try_from("http://[::]:40081")
            .unwrap()
            .connect_with_connector(service_fn(|_: Uri| {
                let path = "/tmp/quote-server.sock";
                UnixStream::connect(path)
            }))
            .await
            .unwrap();

        let mut client = GetQuoteClient::new(channel);

        let request = tonic::Request::new(GetQuoteRequest {
            user_data: "123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678".to_string(),
            nonce: "12345678".to_string(),
        });

        let response = client.get_quote(request).await.unwrap().into_inner();
        assert_eq!(response.quote_type, "TDX");
        assert_ne!(response.quote.len(), 0);
    }

    #[tokio::test]
    #[serial]
    async fn request_to_server_empty_nonce() {
        creat_server().await;

        let channel = Endpoint::try_from("http://[::]:40081")
            .unwrap()
            .connect_with_connector(service_fn(|_: Uri| {
                let path = "/tmp/quote-server.sock";
                UnixStream::connect(path)
            }))
            .await
            .unwrap();

        let mut client = GetQuoteClient::new(channel);

        let request = tonic::Request::new(GetQuoteRequest {
            user_data: "123456781234567812345678123456781234567812345678".to_string(),
            nonce: "".to_string(),
        });

        let response = client.get_quote(request).await.unwrap().into_inner();
        assert_eq!(response.quote_type, "TDX");
        assert_ne!(response.quote.len(), 0);
    }

    #[tokio::test]
    #[serial]
    async fn request_to_server_log_nonce() {
        creat_server().await;

        let channel = Endpoint::try_from("http://[::]:40081")
            .unwrap()
            .connect_with_connector(service_fn(|_: Uri| {
                let path = "/tmp/quote-server.sock";
                UnixStream::connect(path)
            }))
            .await
            .unwrap();

        let mut client = GetQuoteClient::new(channel);

        let request = tonic::Request::new(GetQuoteRequest {
            user_data: "123456781234567812345678123456781234567812345678".to_string(),
            nonce: "123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678123456781234567812345678".to_string(),
        });

        let response = client.get_quote(request).await.unwrap().into_inner();
        assert_eq!(response.quote_type, "TDX");
        assert_ne!(response.quote.len(), 0);
    }

    #[tokio::test]
    #[serial]
    async fn request_to_server_verify_report_data() {
        creat_server().await;

        let channel = Endpoint::try_from("http://[::]:40081")
            .unwrap()
            .connect_with_connector(service_fn(|_: Uri| {
                let path = "/tmp/quote-server.sock";
                UnixStream::connect(path)
            }))
            .await
            .unwrap();

        let mut client = GetQuoteClient::new(channel);

        let request = tonic::Request::new(GetQuoteRequest {
            user_data: "YWJjZGVmZw==".to_string(),
            nonce: "MTIzNDU2Nzg=".to_string(),
        });

        let response = client.get_quote(request).await.unwrap().into_inner();

        let expected_report_data = [
            93, 71, 28, 83, 115, 189, 166, 130, 87, 137, 126, 119, 140, 209, 163, 215, 13, 175,
            225, 101, 64, 195, 196, 202, 15, 37, 166, 241, 141, 49, 128, 157, 164, 132, 67, 50, 9,
            32, 162, 89, 243, 191, 177, 131, 4, 159, 156, 104, 11, 193, 18, 217, 92, 215, 194, 98,
            145, 191, 211, 85, 187, 118, 39, 80,
        ];

        assert_eq!(response.quote_type, "TDX");
        let quote = base64::decode(response.quote.replace("\"", "")).unwrap();
        let mut report_data_in_quote: [u8; 64 as usize] = [0; 64 as usize];
        report_data_in_quote.copy_from_slice(&quote[568..632]);
        assert_eq!(report_data_in_quote, expected_report_data);
    }
}
