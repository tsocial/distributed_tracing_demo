## Setup database

- Start MariaDB docker container

  ```bash
  docker run --name mariadb -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root -d mariadb
  ```

- Create database and user

  ```bash
  docker exec -it mariadb mysql -uroot -proot
  ```

  - Run following SQL commands:

  ```sql
  CREATE SCHEMA IF NOT EXISTS open_census DEFAULT CHARACTER SET utf8;
  CREATE USER IF NOT EXISTS 'admin' IDENTIFIED BY 'password';
  GRANT ALL ON open_census.* TO 'admin';
  \q
  ```

- Access database with the created user

  ```bash
  docker exec -it mariadb mysql open_census -uadmin -ppassword
  ```

  - Some SQL commands:

  ```sql
  SELECT DATABASE();
  SHOW TABLES;
  SELECT DISTINCT table_name, index_name FROM information_schema.statistics WHERE table_schema = 'open_census';
  \q
  ```

## Setup backend services
Using docker for deploying some backend services such as Jaeger (tracing backend), Prometheus (metric backend) or OpenCensus service

### Start Jaeger
```bash
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.6
```
<b>Reference</b>
- [Jaeger Docker All In One](https://www.jaegertracing.io/docs/1.6/getting-started/#all-in-one-docker-image)
- [Getting started](https://www.jaegertracing.io/docs/1.11/getting-started/)

### Start prometheus
- Note that docker volume doesn't work with relative path([here](https://www.quora.com/Do-docker-volumes-not-work-with-relative-paths)) so we use `$PWD` environment variable here.
- Prometheus is pull-based service. We must configure service / interval pulling time configuration in `prometheus.yml`
```bash
docker run -p 9090:9090 -d -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
```
<b>Reference</b>
- [Deploy Prometheus using Docker](https://prometheus.io/docs/prometheus/latest/installation/#using-docker)

### Start OpenCensus service
- Start OpenCensus Agent
TBD

- Start OpenCensus Collector
TBD

## Testing on some backend
For sending request, run following commands:
```bash
curl localhost:3000/login
```


### Test local with zpages.
- Send sample request
- Go to page: http://localhost:3000/debug/tracez

### Test with Jaeger
- Start Jaeger service, enable sending to Jaeger backend when starting server
- Send sample request
- Go to http://localhost:16686/trace

### Test tracing print out to console
- Enable sending to console when starting server
- Send sample request and check on console

Example output
```
#----------------------------------------------
TraceID:      a9034abb689804c1e94b5b2ea1f9cac1
SpanID:       493862a1ce68c244
ParentSpanID: cc4fb688be09e529

Span:    gorm:query
Status:   [0]
Elapsed: 8ms

Annotations:
  detail sql query information
 query=SELECT * FROM `products`  WHERE (`products`.`id` = 1) ORDER BY `products`.`id` ASC LIMIT 1
 variables=[]
 values={%!s(int64=1) tra xanh %!s(int=30000) 2019-05-04 08:30:59 +0700 +07 2019-05-04 08:30:59 +0700 +07}


Attributes:
  - table=products
  - operation=query
  - provider=gorm
```

### Testing prometheus
- Start Prometheus service, enable sending to Prometheus backend when starting server
- Send sample request
- Go to http://localhost:9090/graph for checking metric log
- Go to http://localhost:9090/targets for checking health status

### Testing with OpenCensus Service
TBA

## References:
- [OpenCensus Example](https://github.com/census-instrumentation/opencensus-go/tree/master/examples)
- [OpenCensus Documentation](https://opencensus.io/)
- [How not to measure Latency Talk](https://www.youtube.com/watch?v=lJ8ydIuPFeU)
- [OpenCensus vs OpenTracing](https://github.com/gomods/athens/issues/392)

