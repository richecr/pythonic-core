# PythonicCore

This package provides the core functionality for building the query and connecting to the PythonicSQL database.

## Example

```python
import json
from pythonic_core import dialects, go


class PythonicSQL:
    def __init__(self, dialect: str, uri: str):
        self.client = dialects.new_client(dialect=dialect, uri=uri)


p = PythonicSQL(
    dialect="postgres",
    uri="postgres://default:07YIJKGAkZys@ep-raspy-firefly-a4unwdoj.us-east-1.aws.neon.tech:5432/verceldb?sslmode=require",
)

res = (
    p.client.builder.select(go.Slice_string(["id", "created_at", "action", "coordx"]))
    .from_("interaction")
    .exec()
)

res = json.loads(res.__bytes__())

for r in res:
    print(r["id"], r["action"], r["created_at"], r["coordx"])

```