.PHONY: all

project=victory-frontend
package="softwarehouseio/${project}"
build_tag=`date "+D-%Y-%m-%d%n"`
save_target=softwarehouseio_${project}_latest.tar.gz
runtime_name="running-${build_tag}"
ssh_server=io_softwarehouse_victory
server_workingdir=${project}
local_workdir=${project}_workdir

all :
	echo "options are: \n- release: run all necessary steps to build a release, upload it and restart the server \n- dev: start the local dev environment \n- seeds: build the database from seed files"

release : update build save upload upload-templates upload-locales upload-config-files upload-db upload-cachedb server-restart
	git add victory-frontend/Dockerfile
	git commit -m 'added new release'
	git push origin master

dev :
	cd ${project} && \
	OO_DEBUG="debug" \
	OO_ENV_WORK_DIR="../${local_workdir}" \
	FIREBASE_CONFIG="`pwd`/../${local_workdir}/victory-frontend-firebase-adminsdk.json" \
	OO_ENV_SENDGRID="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" \
	SPMKTOKEN="PqT61ecXA3zdFIiNQSucXfqnL4ko9UBgHRei2L1aB1TnhmJ7mau5rIoCmVKP" \
	GATEWAY_MERCHANT_ID="TEST7008899" \
    GATEWAY_MERCHANT_NAME="FC VICTORY GARMENTS TRADING LLC" \
    GATEWAY_MERCHANT_ADDRESS_LINE1="132 Example Street" \
    GATEWAY_MERCHANT_ADDRESS_LINE2="1234 Example Town" \
	GATEWAY_MASTERCARD_SECRET="F9693FEEA85DF51371D689E2C0B74710" \
	GCLOUD_PROJECT="sh-tt-victory" \
	OO_ENV_PUBLIC_DIR=./public \
	DEBUG=info \
	./node_modules/gulp/bin/gulp.js watch

dev-server :
	cd ${project} && \
	OO_DEBUG="debug" \
	OO_ENV_WORK_DIR="../${local_workdir}" \
	FIREBASE_CONFIG="`pwd`/../${local_workdir}/victory-frontend-firebase-adminsdk.json" \
	OO_ENV_SENDGRID="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" \
	SPMKTOKEN="PqT61ecXA3zdFIiNQSucXfqnL4ko9UBgHRei2L1aB1TnhmJ7mau5rIoCmVKP" \
	GATEWAY_MERCHANT_ID="TEST7008899" \
    GATEWAY_MERCHANT_NAME="FC VICTORY GARMENTS TRADING LLC" \
    GATEWAY_MERCHANT_ADDRESS_LINE1="132 Example Street" \
    GATEWAY_MERCHANT_ADDRESS_LINE2="1234 Example Town" \
	GATEWAY_MASTERCARD_SECRET="F9693FEEA85DF51371D689E2C0B74710" \
	GCLOUD_PROJECT="sh-tt-victory" \
	OO_ENV_PUBLIC_DIR=./public \
	DEBUG=info \
	./node_modules/gulp/bin/gulp.js server:build assets:watch server:spawn browser-sync

seeds :
	cd ${project} && \
	OO_ENV_WORK_DIR="../${local_workdir}" \
	OO_ENV_PUBLIC_DIR=./public \
	FIREBASE_CONFIG="`pwd`/../${local_workdir}/victory-frontend-firebase-adminsdk.json" \
	SPMKTOKEN="PqT61ecXA3zdFIiNQSucXfqnL4ko9UBgHRei2L1aB1TnhmJ7mau5rIoCmVKP" \
	GCLOUD_PROJECT="sh-tt-victory" \
	go run db/seeds/main.go db/seeds/seeds.go

syncstats :
	cd ${project} && \
	OO_ENV_WORK_DIR="../${local_workdir}" \
	OO_ENV_PUBLIC_DIR=./public \
	FIREBASE_CONFIG="`pwd`/../${local_workdir}/victory-frontend-firebase-adminsdk.json" \
	SPMKTOKEN="PqT61ecXA3zdFIiNQSucXfqnL4ko9UBgHRei2L1aB1TnhmJ7mau5rIoCmVKP" \
	GCLOUD_PROJECT="sh-tt-victory" \
	go run db/syncstats/main.go

#dev:
#	echo 'launch a development server build from most recent version and render right away'
#	cd ${project} && OO_ENV_TEMPLATE_DIR="templates" OO_ENV_WORK_DIR="../${project}_workdir" ./node_modules/gulp/bin/gulp.js watch
#
#dev-frontend:
#	echo 'launch a development server build from most recent version and render right away'
#	cd ${project} && OO_ENV_TEMPLATE_DIR="templates" OO_ENV_WORK_DIR="../${project}_workdir" ./node_modules/gulp/bin/gulp.js server:spawn assets:watch browser-sync

update :
	git pull

build-update-gopath-src :
	cd ${project} && \
	tar -cvz --exclude='.git' --exclude='node_modules' --exclude="${project}" \
		-f gopath_src.tar.gz \
		-C ${GOPATH} \
		src

build :
	echo 'building a new release ${project}'
	sed -i -e "s/OO_ENV_VERSION=.*$$/OO_ENV_VERSION=`cd ${project} && git rev-parse HEAD`/g" ./${project}/Dockerfile
	cd ${project} && \
	docker build -t ${package}:${build_tag} . && docker tag ${package}:${build_tag} ${package}:latest

save :
	echo "saving docker image ${package}:latest to /tmp/${save_target}"
	cd ${project} && \
	docker save ${package}:latest | gzip > /tmp/${save_target}

build-dev-env :
	echo "creating the dev env docker image"
	docker build -t victory-dev-env:latest .

start-dev-env :
	echo "starting the dockerized development setup"
	./dev-env/setup.sh
	echo "run: make dev"
	cd dev-env && docker-compose run --rm -p 3000:3000 -p 3001:3001 dev-env

upload :
	echo 'uploading the newest docker image - not completed yet'
	ssh ${ssh_server} "rm ${save_target}  2> /dev/null || echo > /dev/null"
	rsync --progress -avz /tmp/${save_target} ${ssh_server}:~/
	echo 'loading it into the server'
	ssh ${ssh_server} 'cat ~/${save_target} | gunzip | docker load'

upload-templates :
	echo 'uploading templates to ${ssh_server}:/${server_workingdir}'
	rsync --progress -avz ${project}/templates ${ssh_server}:~/${server_workingdir}/

upload-docker-compose :
	rsync --progress -avz ./docker-compose.yml ${ssh_server}:~/

upload-config-files :
	echo 'uploading config files to ${ssh_server}:/${server_workingdir}'
	#rsync --progress -avz ${project}_workdir/config ${ssh_server}:~/${server_workingdir}/
	#echo "upload firebase config file"
	rsync --progress -avz "${local_workdir}/victory-frontend-firebase-adminsdk.json" ${ssh_server}:~/${server_workingdir}/

upload-locales :
	echo 'uploading locales files'
	rsync --progress -avz ${project}/config/locales ${ssh_server}:~/${server_workingdir}/config/

upload-db :
	echo 'uploading database to ${ssh_server}:/${server_workingdir}'
	rsync --progress -avz ${project}_workdir/${project}.db ${ssh_server}:~/${server_workingdir}/

download-db :
	echo 'downloading database from ${ssh_server}:/${server_workingdir}'
	rsync --progress -avz ${ssh_server}:~/${server_workingdir}/${project}.db ${project}_workdir/

upload-images :
	echo 'uploading images to ${ssh_server}:/uploads'
	rsync --progress -avz ${project}_workdir/uploads ${ssh_server}:~/${server_workingdir}/

download-images :
	echo 'downloading images from ${ssh_server}:/uploads'
	rsync --progress -avz ${ssh_server}:~/${server_workingdir}/uploads ${project}_workdir/

upload-nginx-conf :
	echo 'uploading nginx conf file to ${ssh_server}:/nginx/nginx.conf'
	rsync --progress -avz nginx/nginx.conf ${ssh_server}:~/nginx/nginx.conf

upload-cachedb :
	rsync --progress -avz ${project}_workdir/victory-go-cache-boltdb.db ${ssh_server}:~/${server_workingdir}/

download-cachedb :
	rsync --progress -avz ${ssh_server}:~/${server_workingdir}/victory-go-cache-boltdb.db ${project}_workdir/ 

server-restart :
	echo 'stopping and relaunching the server with the latest docker image' 
	ssh ${ssh_server} 'docker-compose stop && docker-compose rm -f && docker-compose up -d'

server-update :
	echo 'updating the server with the latest docker image' 
	ssh ${ssh_server} 'docker-compose up -d'

server-clean-docker :
	ssh ${ssh_server} 'docker rmi $$(docker images -q -f dangling=true)'

server-setup : server-create-workdir upload-templates upload-docker-compose

server-create-workdir :
	ssh ${ssh_server} "mkdir -p ${server_workingdir}  2> /dev/null || echo > /dev/null"

server-clear-cache :
	ssh ${ssh_server} "rm ${server_workingdir}/victory-go-cache-boltdb.db  2> /dev/null || echo > /dev/null"

# sql
sql :
	sqlite3 ${local_workdir}/${project}.db

# tests

#test :
#	cd ${project} && \
#	./node_modules/gulp/bin/gulp.js server:test --cmd=../test.sh

sportmonks-test :
	cd ${project} && \
	OO_DEBUG="debug" \
	OO_ENV_WORK_DIR="../../../${local_workdir}" \
	FIREBASE_CONFIG="`pwd`/../../../${local_workdir}/victory-frontend-firebase-adminsdk.json" \
	OO_ENV_SENDGRID="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" \
	SPMKTOKEN="Mi6unG6iKf2GcQslSV9C2eKEz8L2z5rltIH8KShjn0AF9AZBzhCAMWPeW3pO" \
	GATEWAY_MERCHANT_ID="TEST7008899" \
    GATEWAY_MERCHANT_NAME="FC VICTORY GARMENTS TRADING LLC" \
    GATEWAY_MERCHANT_ADDRESS_LINE1="132 Example Street" \
    GATEWAY_MERCHANT_ADDRESS_LINE2="1234 Example Town" \
	GATEWAY_MASTERCARD_SECRET="F9693FEEA85DF51371D689E2C0B74710" \
	GCLOUD_PROJECT="sh-tt-victory" \
	OO_ENV_PUBLIC_DIR=./public \
	DEBUG=info \
	./node_modules/gulp/bin/gulp.js server:test --cmd="cd libs/gosportmonks/ && /Users/albsen/.gvm/gos/go1.9/bin/go test -cover -v"
