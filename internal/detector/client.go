package detector

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client connects to the Python AI Worker via gRPC.
type Client struct {
	conn    *grpc.ClientConn
	address string
}

// Detection represents a detected object.
type Detection struct {
	ClassName       string            `json:"class_name"`
	Confidence      float32           `json:"confidence"`
	BBoxPixel       [4]float32        `json:"bbox_pixel"` // xmin, ymin, xmax, ymax
	GeoCenter       [2]float64        `json:"geo_center"` // lat, lon
	EstLengthMeters float32           `json:"est_length_m"`
	EstWidthMeters  float32           `json:"est_width_m"`
	Attributes      map[string]string `json:"attributes"`
}

// ChangeRegion represents a detected change area.
type ChangeRegion struct {
	BBox         [4]float32 `json:"bbox"`
	ChangeType   string     `json:"change_type"`
	Significance float32    `json:"significance"`
	GeoCenter    [2]float64 `json:"geo_center"`
}

func NewClient(address string) (*Client, error) {
	return &Client{address: address}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to AI worker at %s: %w", c.address, err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// DetectObjects sends an image to the AI worker for detection.
func (c *Client) DetectObjects(ctx context.Context, imagePath string, targets []string, confidence float32) ([]Detection, error) {
	// TODO: Use generated proto client
	return nil, fmt.Errorf("not yet implemented")
}

// DetectChanges compares two images.
func (c *Client) DetectChanges(ctx context.Context, beforePath, afterPath string, sensitivity float32) ([]ChangeRegion, error) {
	// TODO: Use generated proto client
	return nil, fmt.Errorf("not yet implemented")
}

// Health checks the AI worker status.
func (c *Client) Health(ctx context.Context) (bool, []string, error) {
	// TODO: Use generated proto client
	return false, nil, fmt.Errorf("not yet implemented")
}
