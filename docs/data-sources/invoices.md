---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "scalingo_invoices Data Source - terraform-provider-scalingo"
subcategory: ""
description: |-
  Invoices generated by an account and their details
---

# scalingo_invoices (Data Source)

Invoices generated by an account and their details



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `after` (String) Filter to get all invoices after a given date
- `before` (String) Filter to get all invoices before a given date

### Read-Only

- `id` (String) The ID of this resource.
- `invoices` (List of Object) All invoices returned by the data source (see [below for nested schema](#nestedatt--invoices))

<a id="nestedatt--invoices"></a>
### Nested Schema for `invoices`

Read-Only:

- `billing_month` (String)
- `detailed_items` (List of Object) (see [below for nested schema](#nestedobjatt--invoices--detailed_items))
- `id` (String)
- `invoice_number` (String)
- `items` (List of Object) (see [below for nested schema](#nestedobjatt--invoices--items))
- `pdf_url` (String)
- `state` (String)
- `total_price` (Number)
- `total_price_with_vat` (Number)
- `vat_rate` (Number)

<a id="nestedobjatt--invoices--detailed_items"></a>
### Nested Schema for `invoices.detailed_items`

Read-Only:

- `app` (String)
- `id` (String)
- `label` (String)
- `price` (Number)
- `region` (String)
- `sku` (String)


<a id="nestedobjatt--invoices--items"></a>
### Nested Schema for `invoices.items`

Read-Only:

- `id` (String)
- `label` (String)
- `price` (Number)

