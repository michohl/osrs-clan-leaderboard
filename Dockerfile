# Build the binary
FROM golang:1.25 AS build
COPY . .
RUN go build -o /tmp/osrs-clan-leaderboard

# Copy the binary into a minimal image
FROM gcr.io/distroless/base-debian12
COPY --from=build /tmp/osrs-clan-leaderboard /
CMD ["/osrs-clan-leaderboard"]
