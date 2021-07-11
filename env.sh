#!/bin/sh

if [ ! -f ./config/.env ]
then
    cp ./config/.env.example ./config/.env
    echo "example env copied"
fi