# Use a base image, for example, Ubuntu
FROM ubuntu:22.04

# Set non-interactive environment for apt
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y curl wget 

# Set the working directory inside the container
WORKDIR /app

COPY quote_server /app/quote_server
RUN chmod a+x /app/quote_server

# Run the sleep command for demonstration purposes
CMD ["/app/quote_server"]
