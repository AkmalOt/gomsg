# gomsg

Библиотека на Go для чтения файлов `.msg` (Microsoft Outlook). Позволяет извлекать из письма тему, отправителя, получателей, тело сообщения и вложения.

Файлы `.msg` используют формат OLE2 (CFB). Библиотека парсит этот формат и достаёт из него MAPI-свойства письма.

## Установка

```bash
go get github.com/AkmalOt/gomsg
```

## Использование

### В коде

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/AkmalOt/gomsg"
)

func main() {
    msg, err := gomsg.Open("email.msg")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Тема:", msg.Subject)
    fmt.Println("От:", msg.SenderName, "<"+msg.SenderSMTP+">")
    fmt.Println("Кому:", msg.DisplayTo)
    fmt.Println("Дата:", msg.Date)
    fmt.Println("Текст:", msg.Body)

    // Извлечение вложений
    for _, a := range msg.Attachments {
        fmt.Printf("Вложение: %s (%d байт)\n", a.DisplayName(), a.Size)
        os.WriteFile(a.DisplayName(), a.Data(), 0644)
    }
}
```

### Через командную строку

В комплекте идёт утилита `msgdump`:

```bash
go install github.com/AkmalOt/gomsg/cmd/msgdump@latest
```

```bash
# Сводка по письму
msgdump email.msg

# Вывод в JSON
msgdump -json email.msg

# Извлечь вложения в папку
msgdump -extract ./attachments email.msg

# Вывести тело письма
msgdump -body email.msg

# Вывести заголовки письма
msgdump -headers email.msg
```

Пример вывода:

```
Subject:      Test message
From:         John <john@example.com>
To:           jane@example.com
Date:         2024-01-15 12:30:00 UTC
Class:        IPM.Note
Importance:   Normal

Recipients (1):
  [To] Jane <jane@example.com>

Attachments (1):
  document.pdf (application/pdf, 45230 bytes)
```

## Что можно достать из письма

| Поле | Свойство |
|------|----------|
| Тема | `Message.Subject` |
| Текст (plain) | `Message.Body` |
| Текст (HTML) | `Message.BodyHTML` |
| Имя отправителя | `Message.SenderName` |
| Email отправителя | `Message.SenderSMTP` |
| Кому | `Message.DisplayTo` |
| Копия | `Message.DisplayCC` |
| Скрытая копия | `Message.DisplayBCC` |
| Список получателей | `Message.Recipients` |
| Вложения | `Message.Attachments` |
| Дата отправки | `Message.Date` |
| Дата доставки | `Message.DeliveryTime` |
| Тип сообщения | `Message.MessageClass` |
| Важность | `Message.Importance` |
| Message-ID | `Message.MessageID` |
| Транспортные заголовки | `Message.Headers` |

Также через `Message.Properties` можно получить доступ к любому MAPI-свойству напрямую.

## Вложения

Библиотека извлекает вложенные файлы и поддерживает вложенные `.msg` (письмо внутри письма):

```go
for _, a := range msg.Attachments {
    if a.IsEmbeddedMessage() {
        inner := a.EmbeddedMessage()
        fmt.Println("Вложенное письмо:", inner.Subject)
    } else {
        os.WriteFile(a.DisplayName(), a.Data(), 0644)
    }
}
```

## Кодировки

Поддерживаются строки в UTF-16LE (Unicode) и различных кодовых страницах Windows: 1250-1258, KOI8-R, KOI8-U, ISO-8859, Shift-JIS, EUC-KR, GBK, Big5 и другие. Кодировка определяется автоматически из свойств письма.

## Зависимости

- [richardlehane/mscfb](https://github.com/richardlehane/mscfb) — парсер формата CFB/OLE2
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) — работа с кодировками

## Лицензия

MIT
