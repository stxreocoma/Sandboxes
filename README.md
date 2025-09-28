# Goroutine Pool

Простой и безопасный worker pool для Go.

## Возможности
- Фиксированное число воркеров и очередь задач.
- Безопасный `Submit` с ошибками:  
  - `ErrNilTask` — пустая задача  
  - `ErrQueueFull` — очередь заполнена  
  - `ErrStopped` — пул остановлен
- Корректная остановка через `Stop()`, ждёт активные задачи.
- Хуки:
  - `AfterTask` — вызывается после каждой задачи.
  - `PanicHandler` — обрабатывает паники.

## Пример использования

```go
p := pool.New(
    pool.WithWorkers(2),
    pool.WithQueueSize(2),
    pool.WithAfterTask(func() {
        fmt.Println("after task done")
    }),
    pool.WithPanicHandler(func(where string, r any) {
        fmt.Printf("panic caught in %s: %v\n", where, r)
    }),
)

_ = p.Submit(func() { fmt.Println("task 1") })
_ = p.Submit(func() { fmt.Println("task 2") })

if err := p.Submit(func() {}); errors.Is(err, pool.ErrQueueFull) {
    fmt.Println("expected:", err)
}

_ = p.Stop()