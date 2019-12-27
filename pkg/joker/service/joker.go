package joker_service

import (
	"context"
	"io"
	"os"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	component "github.com/nayotta/metathings/pkg/component"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/nayotta/metathings-component-joker/proto"
)

type JokerService struct {
	module *component.Module
}

func (js *JokerService) get_logger() log.FieldLogger {
	return js.module.Logger()
}

func (js *JokerService) upload_file_streaming(ctx context.Context, req *pb.UploadFileRequest) (*empty.Empty, error) {
	src := req.GetSource().GetValue()
	dst := req.GetDestination().GetValue()

	src_rd, err := os.Open(src)
	if err != nil {
		js.get_logger().WithField("source", src).Errorf("failed to open source")
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	defer src_rd.Close()

	opt, err := component.NewPutObjectStreamingOptionFromPath(src)
	if err != nil {
		js.get_logger().WithField("source", src).Errorf("failed to new put object streaming option from source")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	cancel, errs, err := js.module.PutObjectStreamingWithCancel(dst, src_rd, opt)
	if err != nil {

	}

	if err != nil && err != io.EOF {
		js.get_logger().WithFields(log.Fields{
			"source":      src,
			"destination": dst,
		}).Errorf("failed to put object streaming")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	js.get_logger().WithFields(log.Fields{
		"streaming":   true,
		"source":      src,
		"destination": dst,
	}).Debugf("upload file")

	return &empty.Empty{}, nil
}

func (js *JokerService) upload_file(ctx context.Context, req *pb.UploadFileRequest) (*empty.Empty, error) {
	src := req.GetSource().GetValue()
	dst := req.GetDestination().GetValue()

	src_rd, err := os.Open(src)
	if err != nil {
		js.get_logger().WithField("source", src).Errorf("failed to open source")
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	defer src_rd.Close()

	err = js.module.Kernel().PutObject(dst, src_rd)
	if err != nil {
		js.get_logger().WithFields(log.Fields{
			"source":      src,
			"destination": dst,
		}).Errorf("failed to put object")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	js.get_logger().WithFields(log.Fields{
		"streaming":   false,
		"source":      src,
		"destination": dst,
	}).Debugf("upload file")

	return &empty.Empty{}, nil
}

func (js *JokerService) HANDLE_GRPC_UploadFile(ctx context.Context, in *any.Any) (out *any.Any, err error) {
	req := &pb.UploadFileRequest{}

	if err = ptypes.UnmarshalAny(in, req); err != nil {
		return nil, err
	}

	res, err := js.UploadFile(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err = ptypes.MarshalAny(res)
	if err != nil {
		return nil, err
	}

	return
}

func (js *JokerService) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*empty.Empty, error) {
	streaming := false
	req_streaming := req.GetStreaming()
	if req_streaming != nil {
		streaming = req_streaming.GetValue()
	}

	if streaming {
		return js.upload_file_streaming(ctx, req)
	} else {
		return js.upload_file(ctx, req)
	}
}

func (js *JokerService) DownloadFile(ctx context.Context, req *pb.DownloadFileRequest) (*empty.Empty, error) {
	panic("unimplemented")
}

func (js *JokerService) InitModuleService(m *component.Module) error {
	js.module = m

	return nil
}
