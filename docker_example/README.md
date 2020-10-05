# informatiCup 2021 example solution

## Build

```
docker build --tag icup2021_example .
```

If you rather like to use a pre-built docker image, you can also [pull it from GitHub](https://github.com/orgs/informatiCup/packages/container/package/icup2021_example):

```
docker pull ghcr.io/informaticup/icup2021_example
docker tag ghcr.io/informaticup/icup2021_example icup2021_example
```

## Run

```
docker run -e URL="wss://msoll.de/spe_ed" -e KEY="<Your API key>" icup2021_example
```

## Additional notes
This solution should **only** show the usage of Docker. The selection of programming language, libraries or approaches is open to contestants.
