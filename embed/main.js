---
bindings:
- name: corona
  resource: ':/corona/api/status'
---
"use strict";

var corona = {
    version: '{{ .bindings.corona.version }}',
};
