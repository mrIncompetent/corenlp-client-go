package corenlp_client

//go:generate curl -s -o "corenlp.proto" https://raw.githubusercontent.com/stanfordnlp/CoreNLP/v4.2.0/src/edu/stanford/nlp/pipeline/CoreNLP.proto
//go:generate protoc --go_out=import_path=corenlp_client:. corenlp.proto
