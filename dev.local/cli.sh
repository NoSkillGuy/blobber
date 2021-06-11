#!/bin/bash

root=$(pwd)

echo "
**********************************************
  Welcome to blobber/validator development CLI 
**********************************************
"



echo " "
echo "Please select which blobber/validator you will work on: "

select i in "1" "2" "3" "clean all"; do
    case $i in
        "1"             ) break;;
        "2"             ) break;;
        "3"             ) break;;
        "clean all"     ) rm -rf ./data ;;
    esac
done


install_postgres () {

    echo Installing blobber_postgres in docker...

    [ ! "$(docker ps -a | grep blobber_postgres)" ] && docker run --name blobber_postgres --restart always -p 5432:5432 -e POSTGRES_PASSWORD=postgres -d postgres:11

    [ -d "./data/blobber$i" ] || mkdir -p "./data/blobber$i" 

    echo Initializing database

    [ -d "./data/blobber$i/sql" ] && rm -rf  [ -d "./data/blobber$i/sql" ]

    cp -r ../sql "./data/blobber$i/"
    cd "./data/blobber$i/sql"

    find . -name "*.sql" -exec sed -i '' "s/blobber_user/blobber_user$i/g" {} \;
    find . -name "*.sql" -exec sed -i '' "s/blobber_meta/blobber_meta$i/g" {} \;


    cd $root
    [ -d "./data/blobber$i/bin" ] && rm -rf  [ -d "./data/blobber$i/bin" ]
    cp -r ../bin "./data/blobber$i/"


    cd $root
  
    [ ! "docker ps -a | grep blobber_postgres_init" ] && docker rm blobber_postgres_init --force


    docker run --name blobber_postgres_init \
    --link blobber_postgres:postgres \
    -e  POSTGRES_PORT=5432 \
    -e  POSTGRES_HOST=postgres \
    -e  POSTGRES_USER=postgres  \
    -e  POSTGRES_PASSWORD=postgres \
    -v  $root/data/blobber$i/bin:/blobber/bin \
    -v  $root/data/blobber$i/sql:/blobber/sql \
    postgres:11 bash /blobber/bin/postgres-entrypoint.sh 

    docker rm blobber_postgres_init --force

}

prepareRuntime() {
    echo "Prepare bloober $i: config,files, data, log .."
    cd $root
    [ -d ./data/bloober$i/config ] && rm -rf $root/data/bloober$i/config
    cp -r ../config "./data/bloober$i/"

    cd  ./data/bloober$i/config/

    find . -name "*.yaml" -exec sed -i '' "s/blobber_user/blobber_user$i/g" {} \;
    find . -name "*.yaml" -exec sed -i '' "s/blobber_meta/blobber_meta$i/g" {} \;
    find . -name "*.yaml" -exec sed -i '' "s/postgres/127.0.0.1/g" {} \;
    cd $root/data/bloober$i/

    [ -d files ] || mkdir files
    [ -d data ] || mkdir data
    [ -d log ] || mkdir log
}

start_blobber () {
    echo "Building blobber $i ..."
    cd ../code/go/0chain.net/blobber   
    go build -v -tags "bn256 development" -gcflags="-N -l" -ldflags "-X 0chain.net/core/build.BuildTag=dev" -o $root/data/bloober$i/blobber .

    prepareRuntime;

    echo "Starting blobber $i ..."

    cd $root
    port="505$i"
    grpc_port="703$i"
    hostname="localhost"
    keys_file="../docker.local/keys_config/b0bnode${i}_keys.txt"
    minio_file="../docker.local/keys_config/minio_config.txt"
    config_dir="./data/bloober$i/config"
    files_dir="./data/bloober$i/files"
    log_dir="./data/bloober$i/log"
    db_dir="./data/bloober$i/data"

    ./data/bloober$i/blobber --port $port --grpc_port $grpc_port -hostname $hostname --deployment_mode 0 --keys_file $keys_file  --files_dir $files_dir --log_dir $log_dir --db_dir $db_dir  --minio_file $minio_file --config_dir $config_dir --devserver
}

start_validator () {
    echo "Building validator $i ..."

    cd ../code/go/0chain.net/validator   
    go build -v -tags "bn256 development" -gcflags="-N -l" -ldflags "-X 0chain.net/core/build.BuildTag=dev" -o $root/data/bloober$i/validator .


    prepareRuntime;

    echo "Starting validator $i ..."


    cd $root
    port="506$i"
    hostname="localhost"
    keys_file="../docker.local/keys_config/b0bnode${i}_keys.txt"
    config_dir="./data/bloober$i/config"
    log_dir="./data/bloober$i/log"


    ./data/bloober$i/validator --port $port -hostname $hostname --deployment_mode 0 --keys_file $keys_file  --log_dir $log_dir --config_dir $config_dir --devserver
}

clean () {
    echo "Building blobber $i"

    cd $root

    rm -rf "./data/blobber$i"
}


echo "
**********************************************
            Blobber/Validator $i
**********************************************"

echo " "
echo "Please select what you will do: "

select f in "install postgres" "start blobber" "start validator" "clean"; do
    case $f in
        "install postgres"  )   install_postgres;     break;;
        "start blobber"     )   start_blobber;        break;;
        "start validator"   )   start_validator;      break;;
        "clean"             )   clean;      break;;
    esac
done


