---
bindings:
- name: corona
  resource: ':/corona/api/status'
---
"use strict";

window.corona = {
    version: '{{ .bindings.corona.version }}',
};
