# Protótipo RAG do agente tutor

## Arquitetura resumida

O protótipo implementa RAG sem LangChain, LangGraph ou banco vetorial. O fluxo é:

1. descobrir somente os arquivos permitidos;
2. dividir o Markdown em trechos com sobreposição;
3. gerar embeddings com `gemini-embedding-001`;
4. manter os vetores em memória;
5. comparar a pergunta e os trechos por similaridade de cosseno;
6. retornar até três evidências ou admitir ausência de evidência suficiente.

O limite padrão é `0.65`. Ele continua configurável nas chamadas de busca e é
uma calibração inicial baseada na demonstração: as evidências sobre pagamento
ficaram entre `0.701` e `0.741`, enquanto a pergunta sobre férias produziu um
falso positivo de `0.552`. Esse corte reduz aquele falso positivo, mas não é uma
garantia universal e deve ser recalibrado com mais perguntas representativas.

`rag_tutor.py` contém o núcleo reutilizável. `agente_tutor.py` integra esse núcleo
ao CrewAI. `demo_rag.py` oferece uma demonstração isolada no terminal e não
executa a auditoria nem altera `DIAGNOSTICO.md`.

## Arquivos indexados

A allowlist aceita exclusivamente:

- `README.md` na raiz;
- `DIAGNOSTICO.md` na raiz;
- arquivos com extensão `.md` dentro de `docs/` e seus subdiretórios.

`AGENTS.md` e arquivos Markdown em outros diretórios não são indexados.

## Medidas de segurança

- os caminhos são resolvidos antes da leitura e precisam permanecer dentro da
  raiz do projeto;
- escapes por links simbólicos são rejeitados;
- `.git`, `.env` e variantes como `.env.local` são bloqueados;
- arquivos não Markdown e arquivos fora da allowlist são rejeitados;
- `demo_rag.py` lê somente a entrada `GOOGLE_API_KEY` do `.env` quando ela não
  estiver no ambiente; o RAG nunca indexa ou exibe o `.env` nem imprime a chave;
- o índice existe somente em memória e não grava documentos ou vetores;
- a demonstração funciona em modo somente leitura.

## Executar os testes

No Git Bash para Windows:

```bash
cd /c/Users/marce/projetcs/ada/go-backend/modulo-01/desafio-pedidos
source .venv/Scripts/activate
python -m unittest -v test_rag_tutor.py
```

Se `.venv` não funcionar:

```bash
source venv/Scripts/activate
```

Se o comando `python` não estiver disponível:

```bash
py -3 -m unittest -v test_rag_tutor.py
```

Os testes usam embeddings falsos e não acessam o Gemini.

## Executar a demonstração

Mantenha `GOOGLE_API_KEY` no `.env` da raiz. O script carrega essa configuração
automaticamente, sem exibir seu valor. Uma variável já definida no ambiente tem
prioridade e não é sobrescrita.

```bash
cd /c/Users/marce/projetcs/ada/go-backend/modulo-01/desafio-pedidos
source .venv/Scripts/activate
python demo_rag.py
```

Se `.venv` não funcionar:

```bash
source venv/Scripts/activate
python demo_rag.py
```

Se o comando `python` não estiver disponível:

```bash
py -3 demo_rag.py
```

O programa mostra quantos documentos e trechos foram indexados. Para cada
resultado, exibe arquivo, posição do trecho, similaridade e conteúdo. Também
mostra o limite mínimo usado. Se todos os resultados ficarem abaixo do limite,
exibe `Nenhuma evidência suficiente` e a mensagem formal de ausência.

## Exemplos de perguntas

Cenário com evidência esperada:

```text
Como funciona o fluxo de pagamento de um pedido?
```

Resultado esperado: até três trechos, com scores observados acima de `0.65`,
preservando o nome de cada arquivo de origem.

Cenário sem evidência esperada:

```text
Qual é a política de férias dos funcionários da empresa?
```

Resultado esperado com a calibração atual: nenhum trecho recuperado.

Quando nenhum trecho atingir o limite mínimo, a saída será:

```text
Não encontrei evidência suficiente na documentação indexada.
```

## Como interpretar o resultado

`Arquivo` identifica a origem da evidência. `Posição` começa em 1 e representa a
ordem do trecho dentro desse arquivo. `Similaridade` é o cosseno entre os dois
vetores: quanto mais próximo de 1, maior a proximidade semântica. O score não é
uma probabilidade nem comprova que a documentação está correta.

## Recuperação documental e confirmação no código

O RAG preserva as fontes recuperadas e não decide silenciosamente qual delas é
correta. `README.md` e `DIAGNOSTICO.md` podem representar momentos diferentes
do projeto e conter afirmações potencialmente conflitantes — por exemplo, uma
fonte pode descrever DLQ como existente enquanto outra ainda recomenda sua
implementação. Nesta etapa não há detecção automática de contradições.

Antes de concluir sobre uma funcionalidade, use os trechos apenas para localizar
o assunto e valide a implementação real nos arquivos de código e configuração.

## Limitações conhecidas

- o índice é recriado a cada execução e faz uma chamada ao Gemini por trecho;
- não há cache, persistência, busca híbrida ou reordenação dos resultados;
- o corte de similaridade é heurístico e pode precisar de calibração;
- ainda podem ocorrer falsos positivos acima de `0.65` e falsos negativos abaixo;
- o chunking considera tamanho textual, não tokens nem estrutura semântica;
- a recuperação encontra evidências, mas não gera uma resposta final;
- documentação pode estar desatualizada, contraditória ou refletir outro momento
  do projeto, por isso não substitui a confirmação no código.
