.PHONY: clean test clean-certs clean-server

all: test

.certs/ca.key:
	mkdir -p .certs
	openssl genrsa -out .certs/ca.key 4096

.certs/ca.crt: .certs/ca.key ca.conf
	openssl req -new -x509 -key .certs/ca.key -out .certs/ca.crt -config ca.conf

.certs/server1.key:
	mkdir -p .certs
	openssl genrsa -out .certs/server1.key 4096

.certs/server1.csr: .certs/server1.key server.conf
	openssl req -new -key .certs/server1.key -out .certs/server1.csr -config server.conf

.certs/server1.crt: .certs/ca.crt .certs/ca.key .certs/server1.csr
	openssl x509 -req -in .certs/server1.csr -CA .certs/ca.crt -CAkey .certs/ca.key -CAcreateserial -out .certs/server1.crt -extfile server.conf -extensions req_ext_1

.certs/server2.key:
	mkdir -p .certs
	openssl genrsa -out .certs/server2.key 4096

.certs/server2.csr: .certs/server2.key server.conf
	openssl req -new -key .certs/server2.key -out .certs/server2.csr -config server.conf

.certs/server2.crt: .certs/ca.crt .certs/ca.key .certs/server2.csr
	openssl x509 -req -in .certs/server2.csr -CA .certs/ca.crt -CAkey .certs/ca.key -CAcreateserial -out .certs/server2.crt -extfile server.conf -extensions req_ext_2

.docker-server-running: .certs/server1.crt .certs/server2.crt .certs/server1.key .certs/server2.key Caddyfile
	docker run -d --name caddy \
		-p 8443:8443 \
		-p 9443:9443 \
		-v ${PWD}/Caddyfile:/etc/caddy/Caddyfile \
		-v ${PWD}/.certs/server1.crt:/server1.crt \
		-v ${PWD}/.certs/server1.key:/server1.key \
		-v ${PWD}/.certs/server2.crt:/server2.crt \
		-v ${PWD}/.certs/server2.key:/server2.key \
		caddy
	touch .docker-server-running

test: .docker-server-running
	curl --cacert .certs/ca.crt --resolve local1.test.com:8443:127.0.0.1 https://local1.test.com:8443
	curl --cacert .certs/ca.crt --resolve local2.test.com:9443:127.0.0.1 https://local2.test.com:9443
	curl --cacert .certs/ca.crt --resolve local1.test.com:8888:127.0.0.1 https://local1.test.com:8888
	curl --cacert .certs/ca.crt --resolve local2.test.com:8888:127.0.0.1 https://local2.test.com:8888

clean: clean-certs clean-server

clean-certs:
	rm -rf .certs
	rm -f .srl

clean-server:
	-docker kill caddy
	-docker rm caddy
	rm -f .docker-server-running


