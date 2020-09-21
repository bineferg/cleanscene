.PHONY: scrape clean

# ~~~~~~~~~~ #
# CLEANSCENE #
# ~~~~~~~~~~ #

# Make targets for building and running scripts.

fly:
	go build -o $@ ./cmd/fly
count:
	go build -o $@ ./cmd/count
cleanup:
	go build -o $@ ./cmd/cleanup

stats: 
	go build -o $@ ./cmd/stats

scrape:
	./fly > artist-pages/artist-test.txt

clean-: 
	rm ./artist-pages/*.csv \ rm fly \ rm count \ rm stats
