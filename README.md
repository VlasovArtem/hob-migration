# House Of Bills Migration Script

## How to build

```shell
go build
```

## How to run

Example

```shell
./hob-migration -u http://localhost:3030 -m /path/example.json -i "26522aed-8580-4db1-8de9-2afea0c75550" 
```

### Parameters

* -u, --url - url to HOB (**Required**). Default: `http://localhost:3030`
* -i, --user-id - id of the user registered in HOB (**Required**)
* -m, --migration-path - path to a migration file path

[File example](./example/example.json)

Full Json Example

```json
{
  "groups": "/groups/groups.csv",
  "houses": "/houses/houses.csv",
  "incomes": "/incomes/incomes.csv",
  "payments": "/payments/payments.csv"
}
```

Possible file formats:

- `csv`

## CSV Headers

### Groups

| Name       |
|------------|
| Group Name |

### Houses

| House Identifier                      | Groups                         | Name       | Country                        | City | Address 1      | Address 2      |
|---------------------------------------|--------------------------------|------------|--------------------------------|------|----------------|----------------|
| Reference for the incomes or payments | Group Names (divided by comma) | House Name | Country Code (for example: UA) | City | Address Line 1 | Address Line 2 |

### Incomes

| House Identifier     | Groups                         | Name        | Description        | Date                 | Sum    |
|----------------------|--------------------------------|-------------|--------------------|----------------------|--------|
| Reference to a House | Group Names (divided by comma) | Income Name | Income Description | 2017-12-20T00:00:00Z | 100,01 |

`House Identifier` or `Groups` name requires

### Payments

| House Identifier     | Name        | Description        | Date                 | Sum    |
|----------------------|-------------|--------------------|----------------------|--------|
| Reference to a House | Income Name | Income Description | 2017-12-20T00:00:00Z | 100,01 |

`House Identifier` requires