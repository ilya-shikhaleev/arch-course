Схема решения
![schema](./README.assets/schema.png)

Для запуска выполняем следующий набор команд:

1) Запустите k8s

Запуск миникуба:
```
minikube start --vm-driver=none && kubectl create namespace arch-course && kubens arch-course
```

2) Добавьте namespace arch-course
```
kubectl create namespace arch-course && kubens arch-course
```

3) Запустите приложение
```
make start
```

4) Запустите тесты
```
newman run app_tests.postman_collection.json
```

5) Мониторинги
```
kubectl port-forward service/user-grafana 9000:80
kubectl port-forward service/prometheus-operated 9090
```

6) Запустите нагрузочные тесты на сервис popular
```
make run-stresstest
make stop-stresstest
```