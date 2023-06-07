
## scanRows

- Result is created (what we will return), each entry that is "scanned" will be added to this
  - Created as `&[]T{}`

- For each row, an entry is created.
  - Created as `&[]T{}`


## newFields

- Is passed a reference to the entry created in the last step
  - As in `&&[]T{}`



## Latest Thoughts On Problem

It seems as though, the User that is in the results slice, is a different user to that which is in the resultsFields.obj
which is why calling `fieldsSlice.Append` is referring to a User, but not the User that will be returned (which is 
the one in the results slice).



## PROBLEM LOCATED

The problem definitely occurs during the call to `fieldSlice.append`. My rudimentary debugging has determined that 
at that point, the `fieldSlice.sliceRef` (which points to the objects slice at the start of the function) ends up 
pointing to a different slice than the object's slice by the end of the function.  