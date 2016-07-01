FROM ruby:2.3
RUN apt-get update
RUN apt-get install -y git
RUN wget -O- "http://s3.amazonaws.com/babl/babl-server_linux_amd64.gz" | gunzip > /bin/babl-server && chmod +x /bin/babl-server
RUN gem install thor json activesupport --no-ri --no-rdoc

WORKDIR /tmp
RUN git clone https://7ceab8a099077a03141f45bdb712f038babaee7e@github.com/larskluge/babl-build.git
WORKDIR /tmp/babl-build
RUN cp babl-build /bin/
RUN cp build-config.yml /root/.babl-build.yml
RUN rm -rf /tmp/babl-build
WORKDIR /usr/src

ADD app /bin/
RUN chmod +x /bin/app

CMD ["babl-server"]
