#!/bin/bash

port1=50051
port2=50052
port3=50053

pid1=$(lsof -i :$port1 | grep $port1 | awk '{print $2}')
pid2=$(lsof -i :$port2| grep $port2 | awk '{print $2}')
pid3=$(lsof -i :$port3| grep $port3| awk '{print $2}')

kill -9 "$pid1"
kill -9 "$pid2"
kill -9 "$pid3"


echo "Shutting down all servers"