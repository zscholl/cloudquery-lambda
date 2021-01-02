FROM golang:1.15 as build
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o /main

FROM public.ecr.aws/lambda/go:1
COPY config.yml ${LAMBDA_TASK_ROOT}
COPY --from=build /main ${LAMBDA_TASK_ROOT}
CMD [ "main" ]
