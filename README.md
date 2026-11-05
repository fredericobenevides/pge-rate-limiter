# PGE Rate Limiter — Go, Redis e Strategy Pattern

Projeto implementado com base nos requisitos da atividade **Rate Limiter** do curso Pós Go Expert da FullCycle.

A aplicação implementa um Rate Limiter HTTP em Go, aplicado como middleware. O middleware controla o número de requisições recebidas por endereço IP ou por token de acesso enviado no header `API_KEY`.

O Redis é utilizado para persistir os contadores, janelas de requisição e bloqueios. A arquitetura utiliza o padrão **Strategy**, permitindo substituir o Redis por outra implementação de armazenamento, como uma implementação em memória, sem alterar a lógica do middleware.

## Funcionalidades

- Limitação de requisições por IP.
- Limitação de requisições por token enviado no header `API_KEY`.
- Precedência de token sobre IP.
- Limites e durações configuráveis por variáveis de ambiente.
- Bloqueio temporário após exceder o limite.
- Retorno HTTP `429 Too Many Requests` quando o limite é excedido.
- Persistência com Redis.
- Implementação em memória para testes.
- Testes automatizados para middleware e estratégias de armazenamento.

## Arquitetura

```text
Cliente HTTP
    |
    | GET /...
    | API_KEY: <TOKEN> (opcional)
    v
Rate Limit Middleware
    |
    |-- API_KEY presente?
    |      |
    |      +--> Limitação por Token
    |
    +--> Limitação por IP
    |
    v
StorageStrategy
    |
    |--> RedisStorageStrategy
    |       |
    |       v
    |     Redis
    |
    +--> MemoryStorageStrategy
```

### Componentes

| Componente              | Responsabilidade                                                                      |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `RateLimitMiddleware`   | Identifica IP ou token, chama a estratégia de armazenamento e produz a resposta HTTP. |
| `StorageStrategy`       | Define o contrato para o controle de requisições e persistência.                      |
| `RedisStorageStrategy`  | Implementa a estratégia usando Redis.                                                 |
| `MemoryStorageStrategy` | Implementa a estratégia usando memória                                                |
| Redis                   | Armazena contadores e informações de bloqueio.                                        |

## Fluxo da requisição

1. O cliente envia uma requisição HTTP.
2. O middleware verifica o header `API_KEY`.
3. Quando o token está presente, o limite por token é utilizado.
4. Sem token, o middleware identifica o IP do cliente e aplica o limite por IP.
5. A estratégia verifica o bloqueio, a janela de contagem e o limite de requisições.
6. Se a requisição for permitida, ela segue para o handler da aplicação.
7. Se o limite for excedido, o identificador é bloqueado pelo tempo configurado e a aplicação retorna `429`.

## Resposta ao exceder o limite

Quando o limite é excedido, a aplicação retorna:

```http
HTTP/1.1 429 Too Many Requests
```

Com o corpo:

```text
you have reached the maximum number of requests or actions allowed within a certain time frame
```

Durante o período de bloqueio, todas as novas requisições do mesmo IP ou token continuam recebendo `429`.

## Configuração

As configurações são carregadas por um arquivo `.env` na raiz do projeto.

### Variáveis de ambiente

| Variável                | Descrição                                                                                      | Exemplo      |
| ----------------------- | ---------------------------------------------------------------------------------------------- | ------------ |
| `PORT`                  | Porta HTTP da aplicação.                                                                       | `8080`       |
| `REDIS_HOST`            | Endereço do Redis no formato `host:porta`.                                                     | `redis:6379` |
| `REDIS_STORAGE_ENABLED` | Define se a estratégia Redis é utilizada. Quando `false`, a estratégia em memória é utilizada. | `true`       |
| `WINDOW_DURATION`       | Duração da janela de contagem.                                                                 | `1s`         |
| `REQUEST_IP_LIMIT`      | Máximo de requisições por janela para cada IP.                                                 | `10`         |
| `REQUEST_TOKEN_LIMIT`   | Máximo de requisições por janela para cada token.                                              | `100`        |
| `BLOCK_IP_DURATION`     | Tempo de bloqueio após exceder o limite por IP.                                                | `5m`         |
| `BLOCK_TOKEN_DURATION`  | Tempo de bloqueio após exceder o limite por token.                                             | `5m`         |

> As durações usam o formato aceito por `time.ParseDuration`, como `1s`, `500ms`, `1m` e `5m`.

## Como executar com Docker Compose

Na raiz do projeto:

```bash
docker compose up --build -d
```

Serviços disponíveis:

| Serviço        | Endereço                |
| -------------- | ----------------------- |
| Aplicação HTTP | `http://localhost:8080` |
| Redis          | `localhost:6379`        |

## Como testar

Execute todos os testes:

```bash
go test ./...
```

Para executar testes com detector de condições de corrida:

```bash
go test -race ./...
```

O teste com `-race` é especialmente importante para validar a `MemoryStorageStrategy`, que usa `sync.Mutex` para proteger o estado compartilhado em memória.

## Exemplos de uso

### Requisição limitada por IP

Sem enviar o header `API_KEY`, o limite é aplicado usando o IP do cliente:

```bash
curl -i http://localhost:8080/
```

Para enviar múltiplas requisições:

```bash
for i in {1..15}; do
  curl -i http://localhost:8080/
done
```

### Requisição limitada por token

Envie o token pelo header `API_KEY`:

```bash
curl -i \
  -H 'API_KEY: my-api-key' \
  http://localhost:8080/
```

Para enviar múltiplas requisições usando o token:

```bash
for i in {1..105}; do
  curl -i \
    -H 'API_KEY: my-api-key' \
    http://localhost:8080/
done
```

## Estratégias de persistência

### RedisStorageStrategy

Persistência em Redis, adequada para a aplicação em execução com Docker Compose e para múltiplas instâncias compartilhando o mesmo Redis.

### MemoryStorageStrategy

Armazena os dados em memória usando um `map` protegido por `sync.Mutex`.
