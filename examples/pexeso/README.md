# Pexeso

The [game of pairs](https://en.wikipedia.org/wiki/Concentration_(card_game)), also know as _concentration_, _match match_, or _pexeso_. This one has an extraordinarily beautiful concurrent solution.

![Screenshot](screenshot.png)

The game proceeds like this: You click one card, it turns around. Then you click another card. If the two have the same color, then remain turned. Otherwise, they turn back and you don't see their color. The objective is to turn all the cards.

The concurrent implementation of this game is very neat.

Each card is controlled by a separate goroutine. All cards know a common channel called `pair`. They can both send and receive on this channel. It has type `chan PairMsg`.

```go
type PairMsg struct {
	Color color.Color
	Resp  chan<- bool
}
```

When you click a card, it sends its a `PairMsg` with its color to the `pair` channel. When you click the next card, it receives the message from the `pair` channel, compares the colors and sends `true` to the `Resp` channel if they are the same, otherwise it sends `false`. Then both cards know if they matched or not and will either turn back or stay face up.

A card doesn't know if it's the first or the second one clicked, it simply uses the `select` statement to either send a message or receive it. Then, when two cards are both selecting on the channel, one of them will end up sending and the other one receiving.

Another nice thing about this concurrent implementation is that animations are done using simple for-loops:

```go
for c := 32; c >= 0; c-- {
    env.Draw() <- redraw(float64(c) / 32)
    time.Sleep(time.Second / 32 / 4)
}
```
