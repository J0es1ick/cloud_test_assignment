# HTTP-балансировщик нагрузки на Go
Простой балансировщик нагрузки, написанный на Go, использующий PostgreSQL в качестве базы и Docker.
## Инструкция по запуску
Для начала понадобятся заранее установленные Go и Docker.
```PS
# Клонируем репозиторий
git clone https://github.com/J0es1ick/cloud_test_assignment.git
```
В репозитории изменяем данные в файле config.yaml на необходимые вам.
Структура конфига:
```yaml
server:
  port: "8080"

database:
  host: "host"
  port: "port"
  user: "user"
  password: "password"
  name: "name"
  sslmode: "sslmode"
  connect_timeout: "connect_timeout"

backends:
  - "http://backend1:80"
  - "http://backend2:80"

rate_limit:
  default_capacity: default_capacity
  default_rate: "default_rate"
```
Пример заполнения будет виден в самом репозитории.
Далее билдим и запускаем с помощью докера:
```PS
docker-compose up --build
```
Проверить работоспособность можно посредством следующей команды:
```PS
curl http://localhost:your_port
```
С каждым вызовом она будет прилетать на разные бэки. Для тестов я написал 2 "бэкенда", состоящих из index.html, создать можно и больше.
Если у вас есть Apache Bench, то можно устроить бенчмарк посредством него, используя команду
```PS
ab -n 5000 -c 100 http://localhost:your_port/
```
