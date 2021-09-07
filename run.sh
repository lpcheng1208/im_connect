#!/bin/bash

RUN_ENV='local'
case "$1" in
  'test')
  RUN_ENV='test'
  ;;
  'local')
  RUN_ENV='local'
  ;;
  'prod')
  RUN_ENV='prod'
  ;;
esac


docker build -t im_svc --build-arg envType=$RUN_ENV .
docker rm -f im

case "$1" in
  'test')
  docker run -itd -p 9002:9002 --name=im -v /home/ec2-user/project/im_connect/log:/app/log im_svc
  ;;
  'local')
  docker run -itd -p 9002:9002 --name=im --network my_net im_svc
  ;;
  'prod')
  docker run -itd -p 9002:9002 --name=im -v /home/ec2-user/project/im_connect/log:/app/log im_svc
  ;;
esac
