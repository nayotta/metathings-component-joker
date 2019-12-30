package main

import (
	"context"
	"fmt"
	"path"

	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	cmd_contrib "github.com/nayotta/metathings/cmd/contrib"
	client_helper "github.com/nayotta/metathings/pkg/common/client"
	cmd_helper "github.com/nayotta/metathings/pkg/common/cmd"
	context_helper "github.com/nayotta/metathings/pkg/common/context"
	deviced_pb "github.com/nayotta/metathings/pkg/proto/deviced"
)

type ObjectListOption struct {
	cmd_contrib.ClientBaseOption `mapstructure:",squash"`
	Device                       string
	Module                       string
	Path                         string
	Recursive                    bool
	Depth                        int
}

func NewObjectListOption() *ObjectListOption {
	return &ObjectListOption{
		ClientBaseOption: cmd_contrib.CreateClientBaseOption(),
	}
}

var (
	object_list_opt *ObjectListOption
)

var (
	objectListCmd = &cobra.Command{
		Use:     "list",
		Short:   "list files in simple storage",
		Aliases: []string{"ls"},
		PreRun: cmd_helper.DefaultPreRunHooks(func() {
			if base_opt.Config == "" {
				object_list_opt.BaseOption = *base_opt
			} else {
				cmd_helper.UnmarshalConfig(object_list_opt)
				base_opt = &object_list_opt.BaseOption
			}

			if object_list_opt.Token == "" {
				object_list_opt.Token = cmd_helper.GetTokenFromEnv()
			}
		}),
		RunE: object_list,
	}
)

func GetListOptions() (
	*ObjectListOption,
	cmd_contrib.ServiceEndpointsOptioner,
	cmd_contrib.LoggerOptioner,
) {
	return object_list_opt,
		cmd_contrib.NewServiceEndpointsOptionWithTransportCredentialOption(object_list_opt, object_list_opt),
		object_list_opt
}

func print_objects(objects []*deviced_pb.Object) {
	for _, obj := range objects {
		fmt.Println(path.Join(obj.Prefix, obj.Name))
	}
}

func object_list(cmd *cobra.Command, args []string) error {
	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			GetListOptions,
			cmd_contrib.NewLogger("list"),
			cmd_contrib.NewClientFactory,
		),
		fx.Invoke(func(lc fx.Lifecycle, logger log.FieldLogger, opt *ObjectListOption, cli_fty *client_helper.ClientFactory) {
			lc.Append(fx.Hook{OnStart: func(context.Context) error {
				cli, cfn, err := cli_fty.NewDevicedServiceClient()
				if err != nil {
					return err
				}
				defer cfn()

				ctx := context_helper.WithToken(context.TODO(), object_list_opt.GetToken())

				dir := path.Dir(object_list_opt.Path)
				base := path.Base(object_list_opt.Path)

				req := &deviced_pb.ListObjectsRequest{
					Recursive: &wrappers.BoolValue{Value: object_list_opt.Recursive},
					Depth:     &wrappers.Int32Value{Value: int32(object_list_opt.Depth)},
					Object: &deviced_pb.OpObject{
						Device: &deviced_pb.OpDevice{
							Id: &wrappers.StringValue{Value: object_list_opt.Device},
						},
						Prefix: &wrappers.StringValue{Value: dir},
						Name:   &wrappers.StringValue{Value: base},
					},
				}

				res, err := cli.ListObjects(ctx, req)
				if err != nil {
					return err
				}

				objs := res.GetObjects()

				print_objects(objs)

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
	object_list_opt = NewObjectListOption()

	flags := objectListCmd.Flags()

	flags.StringVar(&object_list_opt.Device, "device", "", "Device ID")
	flags.StringVar(&object_list_opt.Module, "module", "", "Module Name")
	flags.StringVarP(&object_list_opt.Path, "path", "p", "", "List path")
	flags.BoolVar(&object_list_opt.Recursive, "recursive", false, "Recursive query")
	flags.IntVar(&object_list_opt.Depth, "depth", 0, "Maximum depth for recursive query")

	ObjectCmd.AddCommand(objectListCmd)
}
