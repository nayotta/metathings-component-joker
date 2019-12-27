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

type UploadOption struct {
	cmd_contrib.ClientBaseOption `mapstructure:",squash"`
	Device                       string
	Module                       string
	Source                       string
	Destination                  string
	Streaming                    bool
}

func NewUploadOption() *UploadOption {
	return &UploadOption{
		ClientBaseOption: cmd_contrib.CreateClientBaseOption(),
	}
}

var (
	upload_opt *UploadOption
)

var (
	UploadCmd = &cobra.Command{
		Use:     "upload",
		Short:   "upload file form module to simple storage",
		Aliases: []string{"u"},
		PreRun: cmd_helper.DefaultPreRunHooks(func() {
			if base_opt.Config == "" {
				upload_opt.BaseOption = *base_opt
			} else {
				cmd_helper.UnmarshalConfig(upload_opt)
				base_opt = &upload_opt.BaseOption
			}

			if upload_opt.Token == "" {
				upload_opt.Token = cmd_helper.GetTokenFromEnv()
			}
		}),
		RunE: upload,
	}
)

func GetUploadOptions() (
	*UploadOption,
	cmd_contrib.ServiceEndpointsOptioner,
	cmd_contrib.LoggerOptioner,
) {
	return upload_opt,
		cmd_contrib.NewServiceEndpointsOptionWithTransportCredentialOption(upload_opt, upload_opt),
		upload_opt
}

func upload(cmd *cobra.Command, args []string) error {
	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			GetUploadOptions,
			cmd_contrib.NewLogger("upload"),
			cmd_contrib.NewClientFactory,
		),
		fx.Invoke(func(lc fx.Lifecycle, logger log.FieldLogger, opt *UploadOption, cli_fty *client_helper.ClientFactory) {
			lc.Append(fx.Hook{OnStart: func(context.Context) error {
				cli, cfn, err := cli_fty.NewDevicedServiceClient()
				if err != nil {
					return err
				}
				defer cfn()

				ctx := context_helper.WithToken(context.TODO(), upload_opt.GetToken())

				req := pb.UploadFileRequest{
					Streaming:   &wrappers.BoolValue{Value: upload_opt.Streaming},
					Source:      &wrappers.StringValue{Value: upload_opt.Source},
					Destination: &wrappers.StringValue{Value: upload_opt.Destination},
				}
				req_any, err := ptypes.MarshalAny(&req)
				if err != nil {
					return err
				}

				rreq := &deviced_pb.UnaryCallRequest{
					Device: &deviced_pb.OpDevice{
						Id: &wrappers.StringValue{Value: upload_opt.Device},
					},
					Value: &deviced_pb.OpUnaryCallValue{
						Name:   &wrappers.StringValue{Value: upload_opt.Module},
						Method: &wrappers.StringValue{Value: "UploadFile"},
						Value:  req_any,
					},
				}

				_, err = cli.UnaryCall(ctx, rreq)
				if err != nil {
					return err
				}

				logger.WithFields(log.Fields{
					"source":      upload_opt.Source,
					"destination": upload_opt.Destination,
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
	upload_opt = NewUploadOption()

	flags := UploadCmd.Flags()

	flags.StringVar(&upload_opt.Device, "device", "", "Device ID")
	flags.StringVar(&upload_opt.Module, "module", "", "Module Name")
	flags.StringVarP(&upload_opt.Source, "source", "s", "", "Upload Source File to Storage")
	flags.StringVarP(&upload_opt.Destination, "destination", "d", "", "File will save in destination")
	flags.BoolVar(&upload_opt.Streaming, "streaming", true, "Upload File in streaming mode")

	RootCmd.AddCommand(UploadCmd)
}
