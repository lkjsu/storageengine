# storageengine

A Storage engine built in Go.

The inspiration for this project came from working with Data in Enterprise settings. I wanted to know why databases are designed with complex features and understand the tradeoffs made in designing these systems.

What it does?

The storage engine contains three components:

- REPL - Facilitates interaction with the user and currently supports the SELECT and INSERT functions of the engine. It also implements metacommands like .exit.

- Table - A single table has been implemented which includes about 100 pages. Each page is 4096 bytes which is the size read by Operating System from disk. By aligning database pages to that boundary, each database read aligns to exactly one OS read.

What's next ?

- implementing filters
- handling failures, logs

## Key Decisions

- **Fixed-size rows** — simplifies offset calculation, trades flexibility for performance
- **4096 byte pages** — aligns with OS memory pages, minimizes I/O overhead  
- **Binary header** — stores row count so data survives restarts

## Running

```
git clone https://github.com/lkjsu/storageengine
cd storageengine
go run main.go


> insert 1, user, user@email.com
> select
> .exit
```