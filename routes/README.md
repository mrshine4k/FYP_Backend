# Routes directory

This directory is for the specific routing of an object/model in the project, where we specify a route, as well as a delegate(callback) function to handle the request, data is retrieved from the link/request by the colon character `":"`

EX:

```go
func AlbumRoute(router *gin.Engine) {
    //to export function, capitalize first letter
    router.POST("/album", controllers.CreateAlbum())
    router.GET("/album/:albumId", controllers.GetAnAlbum())
    router.PUT("/album/:albumId", controllers.UpdateAnAlbum())
    router.DELETE("/album/:albumId", controllers.DeleteAnAlbum())
    router.GET("/albums", controllers.GetAllAlbums())
}
```
