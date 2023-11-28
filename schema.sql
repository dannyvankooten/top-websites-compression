CREATE TABLE sites (
    url TEXT NOT NULL PRIMARY KEY,
    rank INT NOT NULL,
    checked_at DATETIME NOT NULL,
    compression TEXT NOT NULL,
    size INT NOT NULL,
    size_compressed INT NOT NULL
);
CREATE TABLE sites_tmp (
    rank INT,
    url TEXT
);

.mode csv
.import top-10k-domains.csv sites_tmp

INSERT INTO sites(url, rank, checked_at, compression, size, size_compressed) SELECT url, rank, DATETIME('now', '-1 year') AS checked_at, '' AS compression, 0, 0, 0 FROM sites_tmp;
DROP TABLE sites_tmp;
