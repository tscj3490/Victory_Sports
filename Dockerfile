FROM golang:1.9-alpine3.6 as builder
# 
#       BUILD TIME 

# RUN go-wrapper download   # "go get -d -v ./..."
# RUN go-wrapper install    # "go install -v ./..."
# 
# CMD ["go-wrapper", "run"] # ["get1free"]

RUN apk add git gcc make libc-dev g++ --no-cache

# move stored libs in place
ADD . /go/src/bitbucket.org/softwarehouseio/victory/victory-frontend/

WORKDIR /go/src/bitbucket.org/softwarehouseio/victory/victory-frontend/

RUN tar xfz ./gopath_src.tar.gz -C /go/

RUN go-wrapper download
RUN go-wrapper install

# 
#       RUNTIME

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/victory-frontend /usr/local/bin/victory-frontend
COPY --from=builder /go/src/bitbucket.org/softwarehouseio/victory/victory-frontend/templates /templates
COPY --from=builder /go/src/bitbucket.org/softwarehouseio/victory/victory-frontend/public /public
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

# by default the template dir is the one in the container, for development
# this might make sense to redirect this to be in the workdir which is a
# volume that can easily be changed ...
ENV OO_ENV_PUBLIC_DIR="/public"
ENV OO_ENV_TEMPLATE_DIR="/templates"
ENV OO_ENV_WORK_DIR="/victory-frontend"
ENV OO_ENV_ROOT_DIR="/victory-frontend"
ENV SPMKTOKEN="Mi6unG6iKf2GcQslSV9C2eKEz8L2z5rltIH8KShjn0AF9AZBzhCAMWPeW3pO"
ENV GATEWAY_MERCHANT_ID="TEST7008899"
ENV GATEWAY_MERCHANT_NAME="FC VICTORY GARMENTS TRADING LLC"
ENV GATEWAY_MERCHANT_ADDRESS_LINE1="132 Example Street"
ENV GATEWAY_MERCHANT_ADDRESS_LINE2="1234 Example Town"
ENV GATEWAY_MASTERCARD_SECRET="F9693FEEA85DF51371D689E2C0B74710"
ENV OO_ENV_VERSION=3fedfda77f6f25e89497d51c35e40089e142f098
ENV OO_ENV_SENDGRID="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"

ENV FIREBASE_CONFIG="/victory-frontend/victory-frontend-firebase-adminsdk.json"
ENV GCLOUD_PROJECT="sh-tt-victory"

# make the workdir if it doesnt exist
RUN [ -d ${OO_ENV_WORK_DIR} ] || mkdir ${OO_ENV_WORK_DIR}
# make docker forget anything inside of it and make it easier to mount
# by declaring it a volume
VOLUME [${OO_ENV_WORK_DIR}]
# declare it the workdir
WORKDIR ${OO_ENV_WORK_DIR}

# expose the default port
EXPOSE 8080

CMD ["/usr/local/bin/victory-frontend"]
