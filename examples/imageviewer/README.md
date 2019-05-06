# Image Viewer

This example is a proof of concept that it is actually possible to make real GUI apps with this package :).

![Screenshot](screenshot.png)

The implementation is about 600 LOC, but is very straightforward. Most of the lines are spent on drawing text, buttons, layouts, scrolling, and similar stuff. You are free to use those implementations in your programs, but I didn't include them in the package itself, because they aren't flexible and thought out enough.

There are five elements in the app: three buttons, one file browser, and one viewer (the place where images appear).

All these elements run concurrently and communicate using channels.

The file browser accepts messages from the `cd` channel of type `chan string`. The three buttons send messages to this channel. The _'Dir Up'_ button sends `".."`, the _'Refresh'_ button sends `"."`, and the _'Home'_ button sends the user's home directory.

The viewer element accepts messages from the `view` channel of type `chan string`. When you click two times on a file in the file browser, the browser sends the file name to the viewer over this channel. The viewer will attempt to open and decode the file and shows the image if it succeeds. Otherwise it shows text _'Invalid image'_.

The most visible advantage of the concurrent approach here is that when the image takes longer to load, the rest of the UI remains responsive. Another, subtler advantage is that each element is self-contained, implemented using simple loops. The program should therefore be easy to understand and extend (the only barrier to understanding could be that I wrote it over an evening, but I'll try and make the code cleaner over time).
