package main

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	cmd_contrib "github.com/nayotta/metathings/cmd/contrib"
	client_helper "github.com/nayotta/metathings/pkg/common/client"
	cmd_helper "github.com/nayotta/metathings/pkg/common/cmd"
	context_helper "github.com/nayotta/metathings/pkg/common/context"
	deviced_pb "github.com/nayotta/metathings/pkg/proto/deviced"

	pb "github.com/nayotta/metathings-component-joker/proto"
)

type ObjectUploadOption struct {
	cmd_contrib.ClientBaseOption `mapstructure:",squash"`
	Device                       string
	Module                       string
	Source                       string
	Destination                  string
	Streaming                    bool
}

func NewObjectUploadOption() *ObjectUploadOption {
	return &ObjectUploadOption{
		ClientBaseOption: cmd_contrib.CreateClientBaseOption(),
	}
}

var (
	object_upload_opt *ObjectUploadOption
)

var (
	objectUploadCmd = &cobra.Command{
		Use:     "upload",
		Short:   "upload file form module to simple storage",
		Aliases: []string{"u"},
		PreRun: cmd_helper.DefaultPreRunHooks(func() {
			if base_opt.Config == "" {
				object_upload_opt.BaseOption = *base_opt
			} else {
				cmd_helper.UnmarshalConfig(object_upload_opt)
				base_opt = &object_upload_opt.BaseOption
			}

			if object_upload_opt.Token == "" {
				object_upload_opt.Token = cmd_helper.GetTokenFromEnv()
			}
		}),
		RunE: object_upload,
	}
)

func GetUploadOptions() (
	*ObjectUploadOption,
	cmd_contrib.ServiceEndpointsOptioner,
	cmd_contrib.LoggerOptioner,
) {
	return object_upload_opt,
		cmd_contrib.NewServiceEndpointsOptionWithTransportCredentialOption(object_upload_opt, object_upload_opt),
		object_upload_opt
}

func object_upload(cmd *cobra.Command, args []string) error {
	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			GetUploadOptions,
			cmd_contrib.NewLogger("upload"),
			cmd_contrib.NewClientFactory,
		),
		fx.Invoke(func(lc fx.Lifecycle, logger log.FieldLogger, opt *ObjectUploadOption, cli_fty *client_helper.ClientFactory) {
			lc.Append(fx.Hook{OnStart: func(context.Context) error {
				cli, cfn, err := cli_fty.NewDevicedServiceClient()
				if err != nil {
					return err
				}
				defer cfn()

				ctx := context_helper.WithToken(context.TODO(), object_upload_opt.GetToken())

				req := pb.UploadFileRequest{
					Streaming:   &wrappers.BoolValue{Value: object_upload_opt.Streaming},
					Source:      &wrappers.StringValue{Value: object_upload_opt.Source},
					Destination: &wrappers.StringValue{Value: object_upload_opt.Destination},
				}
				req_any, err := ptypes.MarshalAny(&req)
				if err != nil {
					return err
				}

				rreq := &deviced_pb.UnaryCallRequest{
					Device: &deviced_pb.OpDevice{
						Id: &wrappers.StringValue{Value: object_upload_opt.Device},
					},
					Value: &deviced_pb.OpUnaryCallValue{
						Name:   &wrappers.StringValue{Value: object_upload_opt.Module},
						Method: &wrappers.StringValue{Value: "UploadFile"},
						Value:  req_any,
					},
				}

				_, err = cli.UnaryCall(ctx, rreq)
				if err != nil {
					return err
				}

				logger.WithFields(log.Fields{
					"source":      object_upload_opt.Source,
					"destination": object_upload_opt.Destination,
				}).Infof("upload file")

				return nil
			}})
		}),
	)

	if err := app.Start(context.TODO()); err != nil {
		return err
	}
	if err := app.Err(); err != nil {
		return err
	}

	return nil
}

func init() {
	object_upload_opt = NewObjectUploadOption()

	flags := objectUploadCmd.Flags()

	flags.StringVar(&object_upload_opt.Device, "device", "", "Device ID")
	flags.StringVar(&object_upload_opt.Module, "module", "", "Module Name")
	flags.StringVarP(&object_upload_opt.Source, "source", "s", "", "Upload Source File to Storage")
	flags.StringVarP(&object_upload_opt.Destination, "destination", "d", "", "File will save in destination")
	flags.BoolVar(&object_upload_opt.Streaming, "streaming", true, "Upload File in streaming mode")

	ObjectCmd.AddCommand(objectUploadCmd)
}
