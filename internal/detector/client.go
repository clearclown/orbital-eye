package detector

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/clearclown/orbital-eye/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.DetectorServiceClient
}

func NewClient(address string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024),
			grpc.MaxCallSendMsgSize(100*1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to AI worker at %s: %w", address, err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewDetectorServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) Health(ctx context.Context) (*pb.HealthResponse, error) {
	return c.client.Health(ctx, &pb.HealthRequest{})
}

func (c *Client) DetectFromFile(ctx context.Context, imagePath string, targets []string, confidence float32, gsd float32) (*pb.DetectResponse, error) {
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	return c.client.DetectObjects(ctx, &pb.DetectRequest{
		ImageData:            imgData,
		TargetClasses:        targets,
		ConfidenceThreshold:  confidence,
		GsdMeters:            gsd,
	})
}

func (c *Client) DetectFromPath(ctx context.Context, imagePath string, targets []string, confidence float32, gsd float32, topLat, topLon float64) (*pb.DetectResponse, error) {
	return c.client.DetectObjects(ctx, &pb.DetectRequest{
		ImagePath:           imagePath,
		TargetClasses:       targets,
		ConfidenceThreshold: confidence,
		GsdMeters:           gsd,
		TopLeft:             &pb.GeoPoint{Latitude: topLat, Longitude: topLon},
	})
}

func (c *Client) DetectChangesFromFiles(ctx context.Context, beforePath, afterPath string, sensitivity float32) (*pb.ChangeResponse, error) {
	return c.client.DetectChanges(ctx, &pb.ChangeRequest{
		ImageBeforePath: beforePath,
		ImageAfterPath:  afterPath,
		Sensitivity:     sensitivity,
	})
}
