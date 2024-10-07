CREATE
or REPLACE VIEW logs_view AS
select
    method,
    request_uri,
    request,
    token
from
    logs
order by
    start asc;