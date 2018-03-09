**go-git-server**  
This project is a pure Go implementation of Git server side plumbing operations. Its based on [`go-git`](https://github.com/src-d/go-git) and uses [`GoLevelDB`](https://github.com/syndtr/goleveldb) for storage.  

The `Makefile` has various instructions for linting, testing, building and running.

*Is it production ready?*  
Short answer, no!
This is just a prototype Git implementation based on a custom storage. Its not intended to be production ready. It lacks authentication, which is critical for a real version control server. Even the storage implementation is not optimised for performance and safety.
This project can however be a reference for more complex implementations backed by high performance and high availability storage clusters.

*How to add a custom authentication backend?*  
Authentication is neatly abstracted by `go-git`. Refer to its docs and source code for ideas on how to implemented a custom auth backend.

*What about client side operations in consumer services?*  
`go-git` has several examples for this. The test in this repo also has an example with some in-memory operations like commit, push, pull, etc.
