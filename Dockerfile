FROM golang:1.20-buster AS build

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o /bin/app

FROM debian:buster-slim

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=build /bin/app /bin
COPY .env /bin
COPY /assets /bin/assets/andriod-alarm.mp3

CMD [ "/bin/app" ]