CREATE
OR REPLACE VIEW federations_info AS WITH companies AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        companies
    GROUP BY
        federation_uuid
),
users AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        federation_users
    GROUP BY
        federation_uuid
),
c_users AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        company_users
    GROUP BY
        federation_uuid
),
projects AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        projects
    GROUP BY
        federation_uuid
),
tASks AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        tASks
    GROUP BY
        federation_uuid
),
catalogs AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        catalogs
    GROUP BY
        federation_uuid
),
catalog_data AS (
    SELECT
        federation_uuid,
        count(*) AS c
    FROM
        catalog_data
    GROUP BY
        federation_uuid
)
SELECT
    f.uuid,
    (
        SELECT
            c
        FROM
            companies
        where
            federation_uuid = f.uuid
    ) AS companies,
    (
        SELECT
            c
        FROM
            users
        where
            federation_uuid = f.uuid
    ) AS users,
    (
        SELECT
            c
        FROM
            c_users
        where
            federation_uuid = f.uuid
    ) AS c_users,
    (
        SELECT
            c
        FROM
            projects
        where
            federation_uuid = f.uuid
    ) AS projects,
    (
        SELECT
            c
        FROM
            tASks
        where
            federation_uuid = f.uuid
    ) AS tASks,
    (
        SELECT
            c
        FROM
            catalogs
        where
            federation_uuid = f.uuid
    ) AS catalogs,
    (
        SELECT
            c
        FROM
            catalog_data
        where
            federation_uuid = f.uuid
    ) AS catalog_data
FROM
    federations f;