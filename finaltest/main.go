package main

import (
	//"fmt"
	"context"
	"log"
	"net/http"
	"os"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

)

type Booking struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Room string   `bson:"room" json:"room"`
	Start time.Time   `bson:"start" json:"start"`
	End time.Time   `bson:"end" json:"end"`
}

func wrapError(coll *mongo.Collection, h func(context.Context, *gin.Context, *mongo.Collection) error) func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		err := h(ctx, c, coll)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.AbortWithError(http.StatusNotFound, err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
		}
	}
}

func listBooking(ctx context.Context, coll *mongo.Collection) ([]*Booking, error) {
	cur, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var bookings []*Booking
	for cur.Next(ctx) {
		booking := &Booking{}
		if err := cur.Decode(booking); err != nil {
			return nil, err
		}
		fmt.Println(booking)
		bookings = append(bookings, booking)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}

func listBookingHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	bookings, err := listBooking(ctx, coll)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, bookings)
	return nil
}

func findBooking(ctx context.Context, c *gin.Context, coll *mongo.Collection) (*Booking, error) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	booking := &Booking{}
	err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(booking)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func findBookingHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	booking, err := findBooking(ctx, c, coll)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, booking)
	return nil
}

func addBooking(ctx context.Context, c *gin.Context, coll *mongo.Collection) (*Booking, error) {
	booking := &Booking{}

	if err := c.Bind(&booking); err != nil {
		return nil, err
	}
	result, err := coll.InsertOne(ctx, booking)
	if err != nil {
		return nil, err
	}
	booking.ID = result.InsertedID.(primitive.ObjectID)
	return booking, nil
}

func addBookingHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	booking, err := addBooking(ctx, c, coll)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, booking)
	return nil
}


func removeBooking(ctx context.Context, c *gin.Context, coll *mongo.Collection) (error) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	result, err := coll.DeleteOne(c.Request.Context(), bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		c.Status(http.StatusNotFound)			
	}
	return nil
}

func removeBookingHandler(ctx context.Context, c *gin.Context, coll *mongo.Collection) error {
	err := removeBooking(ctx, c, coll)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()
	//mongoURI := "mongodb://uq0cbxjqm6phgv0fgmeb:eZ3cwUG6wB4vF9OOFeyg@bwwh1wweimeiyct-mongodb.services.clever-cloud.com:27017/bwwh1wweimeiyct"
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("DATABASE_URL")))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	
	//coll := client.Database("bwwh1wweimeiyct").Collection("bookings")
	coll := client.Database(os.Getenv("DATABASE_NAME")).Collection("bookings")
	r := gin.Default()

	r.POST("/bookings", wrapError(coll, addBookingHandler))
	r.GET("/bookings", wrapError(coll, listBookingHandler))
	r.GET("/bookings/:id", wrapError(coll, findBookingHandler))
	r.DELETE("/bookings/:id", wrapError(coll, removeBookingHandler))
	r.Run(":8000")
}
