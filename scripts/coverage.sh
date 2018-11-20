#! /bin/sh

echo "mode: set" > coverage.out;
for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|tools)'`; do
    go test -coverprofile=profile.out -v -count 5 $pkg;
    cat profile.out | grep -v "mode: atomic" >> coverage.out;
done
rm -f profile.out
