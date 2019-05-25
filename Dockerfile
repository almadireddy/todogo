FROM iron/go
WORKDIR /app

ADD todo_linux /app/

ENTRYPOINT ["./todo_linux"]