FROM ruby:2.2

RUN gem install thor
ADD . /babel-build
WORKDIR /babel-build

CMD ["help"]
ENTRYPOINT ["/babel-build/build.rb"]

