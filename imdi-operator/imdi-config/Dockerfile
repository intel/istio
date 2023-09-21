from ubuntu:focal

WORKDIR /

ADD ./imdi-all.yaml imdi-all.yaml
ADD ./samples samples
ADD ./genconfig.sh genconfig.sh
RUN chmod 755 genconfig.sh

ENTRYPOINT ["./genconfig.sh"]
