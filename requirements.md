# Fastest TCP Server Competition

For November, we're going to hold a competition to write the fastest TCP server. All of the information you'll need is below. Please feel free to work in teams! I'll include a link to test software as soon as it's written.

## Instructions

Write a server that adheres to the below requirements. If any of the requirements are NOT met, the submission will be disqualified from the official judging. You’re free to run the code anywhere you’d like; ie vm, docker, cloud, etc.

- [x] Responds on port 3280.
- [x] Take in a 10 digit number larger than 1,000,000, with support for leading 0’s. Ie. 0001000000.
- [x] The server should shutdown if it receives “shutdown” as input.
- [x] The server should kill any connection that sends it malformed data.
- [x] The server should support no more than 6 connections.
- [x] Input should be terminated with a newline sequence.
- [x] Unique entries should be written to a file called “data.0.log” when received and this file should be created or cleared each time the application starts.
- [x] Every 5 seconds, counters should be printed out to STDOUT for the number of unique numbers received, the total for that period, and the total for the duration of time the server has been running. The counters should then be flushed.
- [x] Every 10 seconds, the log should rotate and increment the number in the name, all while only writing unique numbers. Example: data.0.log -> data.1.log -> data.2.log.

## Judging

Judging will be held during the night of the meetup. Benchmarking code will be written, distributed, and ran against servers submitted into the competition. The server with the fastest recorded time, out of 3 runs, will be declared the winner.

Benchmark runs will last 20 seconds from my computer to keep a base expectation on each run.

If you’d like, please take a few moments to describe your approach, testing, anr/or anything interesting learned.

## Prizes

- 1st - $25 Amazon Gift Card
- 2nd - $15 Amazon Gift Card
- 3rd - $10 Amazon Gift Card
