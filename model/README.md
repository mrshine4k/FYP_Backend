# Model directory

This directory is meant for model files, containing a struct, and it's conversion to bson/json format
as well as their validation requirements or not

EX:

```go
type Album struct {
    Id     primitive.ObjectID `bson:"_id,omitempty"`
    Title  string             `bson:"Title,omitempty" validate:"required"`
    Artist string             `bson:"Artist,omitempty" validate:"required"`
    Price  json.Number        `bson:"Price,omitempty" validate:"required"`
}
```
