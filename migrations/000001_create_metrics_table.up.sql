create table metrics(
     id VARCHAR(255) NOT NULL,
     kind VARCHAR(255) NOT NULL,
     delta integer,
     value double precision,
     updated timestamp
);

--Составной индекс т.к. мы поддерживаем уникальность по связке наименование + тип
CREATE UNIQUE INDEX idx_metrics_uniq ON metrics(id, kind);