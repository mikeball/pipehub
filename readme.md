<p align="center">
	<a href="https://pipehub.io"><img src="misc/assets/images/pipe.png" alt="PipeHub" width="500"></a>
</p>
<h3 align="center">A programmable proxy server</h3>

Please, don't use it in production **yet**! It's nowhere near stable and changing too much.

## Overview
PipeHub is a programmable proxy. With it you can insert dynamic logic into the request path.

## Documentation
You can find the complete documentation at https://pipehub.io.

---
o pipe_dynamic mudou de local, foi para:

/internal/application/server/service/pipe/pipe_dynamic.go

mudar o nome pra dynamic ao invés de pipe_dynamic.

---
decidir onde fica o init do pipe para o http.
onde achamos o not found, panic handler, etc...
como colocamos no server?

---
o service pipe inicializa o pipe, mas ele conhece alguma coisa sobre os outros servicos?
tipo, quem vai dar o fetch nos handlers?

talvez devemos ter o HTTP dentro do pipe, e o http fica com toda a lógica para extrair as informações necessárias do pipe no que diz respeito a HTTP, certo?

---
apenas inicializar os pipes que tem host, os que não tem, não devemos fazer nada. tem que ter um if lá no código pra fazer isso.

---
só inicializar o pipe caso tenha algum host pra receber os requests.

---
// func (c *Client) Init() {
// 	c.transport = &http.Transport{
// 		Proxy: http.ProxyFromEnvironment,
// 		DialContext: (&net.Dialer{
// 			Timeout:   30 * time.Second,
// 			KeepAlive: 30 * time.Second,
// 			DualStack: true,
// 		}).DialContext,
// 		MaxIdleConns:          c.MaxIdleConns,
// 		MaxConnsPerHost:       c.MaxConnsPerHost,
// 		MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
// 		IdleConnTimeout:       c.IdleConnTimeout,
// 		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
// 		ExpectContinueTimeout: c.ExpectContinueTimeout,
// 	}
// }

---
payload, err := ioutil.ReadFile(path)
		if err != nil {
			return Config{}, errors.Wrap(err, "load file error")
		}