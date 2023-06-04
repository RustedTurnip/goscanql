# Notes

### Current Task

**Implementing Scanner functionality properly.**

Currently, fields are added to the `fields` entity in three ways:

- A simple field (of primitive type), is added as a "field", for example an int would be added as such, and tracked
  in "referenceFields" as *int (so the original value can be updated).
- A struct is added as a one-to-one. If a field is of type **struct{}, then f.obj will be set to be ***struct{}, and 
  if the struct is set to be struct{}, then it will be added as *struct{}.
- A field that implements the Scanner interface is added seperately.