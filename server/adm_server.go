package main

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
)

type Conta struct {
	ID    int
	Saldo float64
	Mutex sync.Mutex
}

type ServicoContas struct {
	contas            map[int]*Conta
	mutex             sync.Mutex
	transacoes        map[string]bool
	contadorTransacao int
	contadorConta     int
}

type DepositoArgs struct {
	IDConta     int
	Valor       float64
	TransacaoID string
}

type SacarArgs struct {
	IDConta     int
	Valor       float64
	TransacaoID string
}

func (s *ServicoContas) AbrirConta(saldoInicial float64, resultado *int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	novaContaID := s.contadorConta
	s.contadorConta++

	s.contas[novaContaID] = &Conta{
		ID:    novaContaID,
		Saldo: saldoInicial,
	}
	*resultado = novaContaID

	fmt.Printf("Conta criada com sucesso! ID da conta: %d, Saldo inicial: %.2f\n", novaContaID, saldoInicial)
	return nil
}

func (s *ServicoContas) Sacar(args SacarArgs, resultado *bool) error {
	s.mutex.Lock()
	if s.transacoes[args.TransacaoID] {
		s.mutex.Unlock()
		*resultado = true
		return nil
	}
	conta, existe := s.contas[args.IDConta]
	s.mutex.Unlock()

	if !existe {
		*resultado = false
		return fmt.Errorf("Conta não encontrada")
	}

	conta.Mutex.Lock()
	defer conta.Mutex.Unlock()

	if conta.Saldo < args.Valor {
		*resultado = false
		return fmt.Errorf("saldo insuficiente")
	}

	// Realiza o saque
	conta.Saldo -= args.Valor

	s.mutex.Lock()
	s.transacoes[args.TransacaoID] = true
	s.mutex.Unlock()

	*resultado = true
	fmt.Printf("Saque realizado com sucesso! Conta ID: %d, Valor: %.2f, Transação ID: %s\n", args.IDConta, args.Valor, args.TransacaoID)
	return nil
}

func (s *ServicoContas) Depositar(args DepositoArgs, resultado *bool) error {
	// Inicia uma seção crítica para verificar e marcar a transação
	s.mutex.Lock()
	if s.transacoes[args.TransacaoID] { // Verifica se a transação já foi processada
		s.mutex.Unlock() // Se a transação já foi processada, libera o lock
		*resultado = true
		return nil
	}

	s.transacoes[args.TransacaoID] = true
	conta, existe := s.contas[args.IDConta]
	s.mutex.Unlock()

	if !existe {
		*resultado = false
		return fmt.Errorf("Conta não encontrada")
	}

	conta.Mutex.Lock()
	defer conta.Mutex.Unlock()

	conta.Saldo += args.Valor

	*resultado = true
	fmt.Printf("Depósito realizado com sucesso! Conta ID: %d, Valor: %.2f, Transação ID: %s\n", args.IDConta, args.Valor, args.TransacaoID)
	return nil
}

func (s *ServicoContas) FecharConta(idConta int, resultado *bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conta, existe := s.contas[idConta]
	if !existe {
		*resultado = false
		return fmt.Errorf("Conta não encontrada")
	}

	conta.Mutex.Lock()
	defer conta.Mutex.Unlock()

	delete(s.contas, idConta)
	*resultado = true

	fmt.Printf("Conta fechada com sucesso! Conta ID: %d\n", idConta)
	return nil
}

func (s *ServicoContas) ConsultarSaldo(idConta int, resultado *float64) error {
	s.mutex.Lock()
	conta, existe := s.contas[idConta]
	s.mutex.Unlock()
	if !existe {
		return fmt.Errorf("Conta não encontrada")
	}

	conta.Mutex.Lock()
	defer conta.Mutex.Unlock()
	*resultado = conta.Saldo
	return nil
}

func (s *ServicoContas) LimparDados(_ struct{}, _ *struct{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.contas = make(map[int]*Conta)
	s.transacoes = make(map[string]bool)
	s.contadorConta = 1
	s.contadorTransacao = 1

	fmt.Println("Dados do servidor limpos com sucesso!")
	return nil
}

func main() {
	servico := &ServicoContas{
		contas:            make(map[int]*Conta),
		transacoes:        make(map[string]bool),
		contadorTransacao: 1,
		contadorConta:     1,
	}

	rpc.Register(servico)
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
