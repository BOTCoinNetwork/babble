package babble

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/mosaicnetworks/babble/src/common"
	"github.com/mosaicnetworks/babble/src/crypto"
	h "github.com/mosaicnetworks/babble/src/hashgraph"
	"github.com/mosaicnetworks/babble/src/net"
	"github.com/mosaicnetworks/babble/src/node"
	"github.com/mosaicnetworks/babble/src/peers"
	"github.com/mosaicnetworks/babble/src/service"
	"github.com/sirupsen/logrus"
)

type Babble struct {
	Config    *BabbleConfig
	Node      *node.Node
	Transport net.Transport
	Store     h.Store
	Peers     *peers.PeerSet
	Service   *service.Service
}

func NewBabble(config *BabbleConfig) *Babble {
	engine := &Babble{
		Config: config,
	}

	return engine
}

func (b *Babble) initTransport() error {
	transport, err := net.NewTCPTransport(
		b.Config.BindAddr,
		nil,
		b.Config.MaxPool,
		b.Config.NodeConfig.TCPTimeout,
		b.Config.NodeConfig.JoinTimeout,
		b.Config.Logger,
	)

	if err != nil {
		return err
	}

	b.Transport = transport

	return nil
}

func (b *Babble) initPeers() error {
	if !b.Config.LoadPeers {
		if b.Peers == nil {
			return fmt.Errorf("Did not load peers but none was present")
		}

		return nil
	}

	peerStore := peers.NewJSONPeerSet(b.Config.DataDir)

	participants, err := peerStore.PeerSet()

	if err != nil {
		return err
	}

	b.Peers = participants

	return nil
}

func (b *Babble) initStore() error {
	if !b.Config.Store {
		b.Config.Logger.Debug("Creating InmemStore")
		b.Store = h.NewInmemStore(b.Config.NodeConfig.CacheSize)
	} else {
		b.Config.Logger.WithField("path", b.Config.BadgerDir()).Debug("BadgerDB")

		bootstrap := b.Config.NodeConfig.Bootstrap
		dbpath := b.Config.BadgerDir()
		i := 1

		for {
			if _, err := os.Stat(dbpath); err == nil {
				b.Config.Logger.Debugf("%s already exists", dbpath)

				if bootstrap {
					break
				}

				dbpath = fmt.Sprintf("%s(%d)", b.Config.BadgerDir(), i)
				b.Config.Logger.Debug("No Bootstrap - using new db %s", dbpath)
				i++
			} else {
				break
			}
		}

		b.Config.Logger.WithField("path", dbpath).Debug("Creating BadgerStore")

		dbStore, err := h.NewBadgerStore(b.Config.NodeConfig.CacheSize, dbpath)
		if err != nil {
			return err
		}

		b.Store = dbStore
	}

	return nil
}

func (b *Babble) initKey() error {
	if b.Config.Key == nil {
		pemKey := crypto.NewPemKey(b.Config.DataDir)

		privKey, err := pemKey.ReadKey()

		if err != nil {
			b.Config.Logger.Warn("Cannot read private key from file", err)

			privKey, err = Keygen(b.Config.DataDir)

			if err != nil {
				b.Config.Logger.Error("Cannot generate a new private key", err)

				return err
			}

			pem, _ := crypto.ToPemKey(privKey)

			b.Config.Logger.Info("Created a new key:", pem.PublicKey)
		}

		b.Config.Key = privKey
	}
	return nil
}

func (b *Babble) initNode() error {
	key := b.Config.Key
	pub := crypto.FromECDSAPub(&key.PublicKey)
	id := common.Hash32(pub)
	moniker := b.Config.Moniker

	p, ok := b.Peers.ByID[id]
	if ok {
		if p.Moniker != moniker {
			b.Config.Logger.WithFields(logrus.Fields{
				"json_moniker": p.Moniker,
				"cli_moniker":  moniker,
			}).Debugf("Using moniker from peers.json file")
			moniker = p.Moniker
		}
	}

	b.Config.Logger.WithFields(logrus.Fields{
		"participants": b.Peers,
		"id":           id,
		"moniker":      moniker,
	}).Debug("PARTICIPANTS")

	b.Node = node.NewNode(
		&b.Config.NodeConfig,
		id,
		key,
		moniker,
		b.Peers,
		b.Store,
		b.Transport,
		b.Config.Proxy,
	)

	if err := b.Node.Init(); err != nil {
		return fmt.Errorf("failed to initialize node: %s", err)
	}

	return nil
}

func (b *Babble) initService() error {
	if b.Config.ServiceAddr != "" {
		b.Service = service.NewService(b.Config.ServiceAddr, b.Node, b.Config.Logger)
	}
	return nil
}

func (b *Babble) Init() error {
	if b.Config.Logger == nil {
		b.Config.Logger = logrus.New()
	}

	if err := b.initPeers(); err != nil {
		return err
	}

	if err := b.initStore(); err != nil {
		return err
	}

	if err := b.initTransport(); err != nil {
		return err
	}

	if err := b.initKey(); err != nil {
		return err
	}

	if err := b.initNode(); err != nil {
		return err
	}

	if err := b.initService(); err != nil {
		return err
	}

	return nil
}

func (b *Babble) Run() {
	if b.Service != nil {
		go b.Service.Serve()
	}

	b.Node.Run(true)
}

func Keygen(datadir string) (*ecdsa.PrivateKey, error) {
	pemKey := crypto.NewPemKey(datadir)

	_, err := pemKey.ReadKey()

	if err == nil {
		return nil, fmt.Errorf("Another key already lives under %s", datadir)
	}

	privKey, err := crypto.GenerateECDSAKey()

	if err != nil {
		return nil, err
	}

	if err := pemKey.WriteKey(privKey); err != nil {
		return nil, err
	}

	return privKey, nil
}
