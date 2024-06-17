# jackett-stremio
A StremIO addon for Jackett

## TODO

- [x] Generate torrentID & cache via torrentID
- [x] Use torrentID & fileID in the download link
- [ ] Use a better algorithm to search for variants in realdebrid.GetFiles
- [ ] Support uncached torrents
- [ ] Cache redirect/download link?
- [ ] Consider buffering for better batching & shuffling
- [x] Support exclusions (Remux, CAM) ...
- [x] Check if it matches IMDB
- [x] Only download necessary files
- [x] Identify fileID by file size
- [x] Identify fileID by file type
- [x] Prioritize matching by IMDB
- [ ] Should loop through all pages in torrents to search for hash
- [ ] Check why couldn't find episode in House S08 1080p BluRay x265 RARBG ORARBG
- [x] Deduplicate infoHash
- [ ] Check flaky results with House S08E01
- [ ] Better pattern to locate files for series
- [ ] Parse season only from the torrent title
