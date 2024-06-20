# prowlarr-stremio
A StremIO addon for Prowlarr

## TODO

- [x] Generate torrentID & cache via torrentID
- [x] Use torrentID & fileID in the download link
- [x] Use a better algorithm to search for variants in realdebrid.GetFiles. Merge files?
- [ ] Support uncached torrents
- [x] Cache redirect/download link?
- [x] Consider buffering for better batching & shuffling. Note: Buffering didn't improve much.
- [x] Support exclusions (Remux, CAM) ...
- [x] Check if it matches IMDB
- [x] Only download necessary files
- [x] Identify fileID by file size
- [x] Identify fileID by file type
- [x] Prioritize matching by IMDB in results
- [ ] Min requirements good candidates
- [ ] Should loop through all pages in torrents to search for hash
- [ ] Check why couldn't find episode in House S08 1080p BluRay x265 RARBG ORARBG
- [x] Deduplicate infoHash
- [x] Check flaky results with House S08E01
- [ ] Better pattern to locate files for series
- [x] Parse season only from the torrent title
- [ ] Merge torrent info when deduplicating
- [x] Enhance stremio APIs with caching
- [x] Forward IP to realdebrid
- [ ] Different strategy to forward IP Address
- [ ] Parse Multi Year
- [x] Cache infoHash instead of magnetURI
- [x] Support /configure and userData
- [x] Move prowlarr to userData
- [ ] Fix checking RD cache not showing full paths. Should move season/episode/ to path instead of fileID
- [ ] Parse HDR
- [x] Support timeout
- [ ] Smarter source to support pagination
