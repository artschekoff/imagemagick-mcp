package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const serverName = "imagemagick-mcp"

func serve() error {
	s := server.NewMCPServer(serverName, version,
		server.WithToolCapabilities(false),
	)

	s.AddTool(
		mcp.NewTool("describe_imagemagick_interface",
			mcp.WithDescription("Returns server metadata and available tools"),
		),
		describeHandler,
	)

	s.AddTool(
		mcp.NewTool("crop_resize_blur_bg",
			mcp.WithDescription("Resize/crop image to exact dimensions. If aspect ratio differs, use mode: blur|cover|contain."),
			mcp.WithString("input_path",
				mcp.Required(),
				mcp.Description("Absolute path to the source image"),
			),
			mcp.WithString("output_path",
				mcp.Required(),
				mcp.Description("Absolute path for the output image (parent dirs created automatically)"),
			),
			mcp.WithNumber("width",
				mcp.Required(),
				mcp.Description("Target width in pixels"),
			),
			mcp.WithNumber("height",
				mcp.Required(),
				mcp.Description("Target height in pixels"),
			),
			mcp.WithString("mode",
				mcp.Description("Resize mode: blur (default), cover, or contain"),
			),
			mcp.WithNumber("blur",
				mcp.Description("Blur radius for blur mode (default: 30)"),
			),
		),
		cropResizeBlurBgHandler,
	)

	return server.ServeStdio(s)
}

func describeHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(
		`{"server":"imagemagick-mcp","tools":"crop_resize_blur_bg","note":"Uses ImageMagick. If aspect ratio differs, pads with blurred version of the image."}`,
	), nil
}

func cropResizeBlurBgHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	inputPath, _ := args["input_path"].(string)
	outputPath, _ := args["output_path"].(string)

	if inputPath == "" {
		return mcp.NewToolResultError("input_path is required"), nil
	}
	if outputPath == "" {
		return mcp.NewToolResultError("output_path is required"), nil
	}

	widthF, ok := args["width"].(float64)
	if !ok {
		return mcp.NewToolResultError("width is required and must be a number"), nil
	}
	heightF, ok := args["height"].(float64)
	if !ok {
		return mcp.NewToolResultError("height is required and must be a number"), nil
	}

	width := int(widthF)
	height := int(heightF)

	mode := "blur"
	if m, ok := args["mode"].(string); ok && m != "" {
		mode = m
	}

	blurRadius := 30
	if b, ok := args["blur"].(float64); ok {
		blurRadius = int(b)
	}

	if err := cropResizeBlurBg(inputPath, outputPath, width, height, mode, blurRadius); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf(`{"ok":"true","output_path":%q,"mode":%q}`, outputPath, mode),
	), nil
}
