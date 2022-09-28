# build stage with golang image
FROM golang:1.17-alpine AS build
WORKDIR /demoapp
COPY go.mod *go.sum *.go ./
RUN CGO_ENABLED=0 go build -o ./demoapp

# app stage with linux base image
FROM alpine
WORKDIR /demoapp
COPY assets/ ./assets/
COPY templates/ ./templates/
COPY --from=build /demoapp/demoapp ./
ENV PORT=8080
EXPOSE ${PORT}
CMD [ "./demoapp" ]
