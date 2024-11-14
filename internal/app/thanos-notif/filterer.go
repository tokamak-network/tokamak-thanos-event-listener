package thanosnotif

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/tokamak-network/tokamak-thanos-event-listener/pkg/log"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindings"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/bindingspreview"
	"github.com/tokamak-network/tokamak-thanos/op-bindings/predeploys"
)

func (p *App) getOptimismPortalFilterers() (optimismPortalFilterer *bindingspreview.OptimismPortal2Filterer, err error) {
	optimismPortalFilterer, err = bindingspreview.NewOptimismPortal2Filterer(common.HexToAddress(p.cfg.OptimismPortal), p.l1Client.GetClient())
	if err != nil {
		log.GetLogger().Errorw("OptimismPortalFilterer instance fail", "error", err)
		return nil, err
	}
	return optimismPortalFilterer, nil
}

func (p *App) getL2ToL1MessagePasserFilterers() (l2ToL1MessagePasserFilterer *bindings.L2ToL1MessagePasserFilterer, err error) {
	l2ToL1MessagePasserFilterer, err = bindings.NewL2ToL1MessagePasserFilterer(common.HexToAddress(predeploys.L2ToL1MessagePasser), p.l2Client.GetClient())
	if err != nil {
		log.GetLogger().Errorw("L2ToL1MessagePasserFilterer instance fail", "error", err)
		return nil, err
	}
	return l2ToL1MessagePasserFilterer, nil
}

func (p *App) getBridgeFilterers() (l1BridgeFilterer *bindings.L1StandardBridgeFilterer, l2BridgeFilterer *bindings.L2StandardBridgeFilterer, err error) {
	l1BridgeFilterer, err = bindings.NewL1StandardBridgeFilterer(common.HexToAddress(p.cfg.L1StandardBridge), p.l1Client.GetClient())
	if err != nil {
		log.GetLogger().Errorw("L1StandardBridgeFilterer instance fail", "error", err)
		return nil, nil, err
	}

	l2BridgeFilterer, err = bindings.NewL2StandardBridgeFilterer(common.HexToAddress(p.cfg.L2StandardBridge), p.l2Client.GetClient())
	if err != nil {
		log.GetLogger().Errorw("L2StandardBridgeFilterer instance fail", "error", err)
		return nil, nil, err
	}

	return l1BridgeFilterer, l2BridgeFilterer, nil
}

func (p *App) getUSDCBridgeFilterers() (l1UsdcBridgeFilterer *bindings.L1UsdcBridgeFilterer, l2UsdcBridgeFilterer *bindings.L2UsdcBridgeFilterer, err error) {
	l1UsdcBridgeFilterer, err = bindings.NewL1UsdcBridgeFilterer(common.HexToAddress(p.cfg.L1UsdcBridge), p.l1Client.GetClient())
	if err != nil {
		log.GetLogger().Errorw("Failed to init the L1UsdcBridgeFilterer", "error", err)
		return nil, nil, err
	}

	l2UsdcBridgeFilterer, err = bindings.NewL2UsdcBridgeFilterer(common.HexToAddress(p.cfg.L2UsdcBridge), p.l2Client.GetClient())
	if err != nil {
		log.GetLogger().Errorw("Failed to init the L2UsdcBridgeFilterer", "error", err)
		return nil, nil, err
	}

	return l1UsdcBridgeFilterer, l2UsdcBridgeFilterer, nil
}
