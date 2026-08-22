# PostgreSQL deployment

The compose stack uses PostgreSQL 17. Apply `migrations` in lexical order before serving traffic. Both migrations and demo inserts are repeatable.
