
Тестирование

CMD
0) Создать подписку
curl -X POST "http://localhost:8080/api/v1/subscriptions/" -H "Content-Type: application/json" -d "{\"service_name\":\"Yandex Plus\",\"price\":400,\"user_id\":\"60601fee-2bf1-4721-ae6f-7636e79a0cba\",\"start_date\":\"07-2025\"}"

нагляднее:
curl -X POST "http://localhost:8080/api/v1/subscriptions/" ^
  -H "Content-Type: application/json" ^
  -d "{
    \"service_name\": \"Yandex Plus\",
    \"price\": 400,
    \"user_id\": \"60601fee-2bf1-4721-ae6f-7636e79a0cba\",
    \"start_date\": \"07-2025\"
  }"

ответ:
{"id":"37832fef-da25-4a57-a68a-196ce116153f","service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025","created_at":"2025-12-15T12:54:39.311017+03:00","updated_at":"2025-12-15T12:54:39.311017+03:00"}

1) GET по id
curl "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/"

2) LIST (список)
Все:
curl "http://localhost:8080/api/v1/subscriptions/?limit=10&offset=0"

Фильтр по user_id:
curl "http://localhost:8080/api/v1/subscriptions/?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"

3) SUMMARY (агрегация за период)

Для одной подписки 400 руб/мес и периода 07-2025..10-2025 должно быть 4 месяца = 1600:
curl "http://localhost:8080/api/v1/subscriptions/summary?from=07-2025&to=10-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"
Ожидаемый ответ:  {"total":1600}

4) UPDATE (изменить цену)
curl -X PUT "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/" -H "Content-Type: application/json" -d "{\"price\":500}"

Потом снова SUMMARY за тот же период — будет 2000.

5) DELETE
curl -X DELETE "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/"

проверка
curl -i "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/"
итог 404.

6) Проверка фидбека “update несуществующей”
curl -i -X PUT "http://localhost:8080/api/v1/subscriptions/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/" -H "Content-Type: application/json" -d "{\"price\":123}"
Правильное поведение: 404,  не 400/500.

---------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------

PowerShell
1) Создать подписку

$body = @{
    service_name = "Yandex Plus"
    price        = 400
    user_id      = "60601fee-2bf1-4721-ae6f-7636e79a0cba"
    start_date   = "07-2025"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/subscriptions/" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body


2 GET по id 

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/" `
  -Method Get

3 LIST 

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/subscriptions/?limit=10&offset=0" `
  -Method Get

4 SUMMARY 

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/subscriptions/summary?from=07-2025&to=10-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba" `
  -Method Get

Ответ :
$total = (Invoke-RestMethod ...).total

5 UPDATE

$body = @{
    price = 500
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/" `
  -Method Put `
  -ContentType "application/json" `
  -Body $body

6 DELETE
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/subscriptions/37832fef-da25-4a57-a68a-196ce116153f/" `
  -Method Delete