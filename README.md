# Rodar o servidor
go run server/adm_server.go

# Rodar os codigos nos clientes agencia
go run agencia/agencia_client.go localhost abrir 100
go run agencia/agencia_client.go localhost depositar 1 100 
go run agencia/agencia_client.go localhost sacar 1 20  
go run agencia/agencia_client.go localhost saldo 1 
go run agencia/agencia_client.go localhost fechar 1

# Rodar os codigos nos clientes caixia
go run caixa/caixa_client.go localhost depositar 1 100 
go run caixa/caixa_client.go localhost sacar 1 20  
go run caixa/caixa_client.go localhost saldo 1    

# Rodar os testes
cd tests
go test

