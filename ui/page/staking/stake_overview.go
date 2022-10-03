package staking

import (
	"context"
	"fmt"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil/v4"
	"gitlab.com/raedah/cryptopower/app"
	"gitlab.com/raedah/cryptopower/libwallet/wallets/dcr"
	"gitlab.com/raedah/cryptopower/listeners"
	"gitlab.com/raedah/cryptopower/ui/cryptomaterial"
	"gitlab.com/raedah/cryptopower/ui/load"
	"gitlab.com/raedah/cryptopower/ui/modal"
	"gitlab.com/raedah/cryptopower/ui/page/components"
	tpage "gitlab.com/raedah/cryptopower/ui/page/transaction"
	"gitlab.com/raedah/cryptopower/ui/values"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

const OverviewPageID = "staking"

type Page struct {
	*load.Load
	// GenericPageModal defines methods such as ID() and OnAttachedToNavigator()
	// that helps this Page satisfy the app.Page interface. It also defines
	// helper methods for accessing the PageNavigator that displayed this page
	// and the root WindowNavigator.
	*app.GenericPageModal
	*listeners.TxAndBlockNotificationListener

	list *widget.List

	ctx       context.Context // page context
	ctxCancel context.CancelFunc

	tickets []*transactionItem

	ticketOverview *dcr.StakingOverview

	ticketsList   *cryptomaterial.ClickableList
	stakeSettings *cryptomaterial.Clickable
	stake         *cryptomaterial.Switch
	infoButton    cryptomaterial.IconButton

	ticketPrice  string
	totalRewards string
}

func NewStakingPage(l *load.Load) *Page {
	pg := &Page{
		Load:             l,
		GenericPageModal: app.NewGenericPageModal(OverviewPageID),
	}

	pg.list = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}

	pg.ticketOverview = new(dcr.StakingOverview)

	pg.initStakePriceWidget()
	pg.initTicketList()

	return pg
}

// OnNavigatedTo is called when the page is about to be displayed and
// may be used to initialize page features that are only relevant when
// the page is displayed.
// Part of the load.Page interface.
func (pg *Page) OnNavigatedTo() {
	// pg.ctx is used to load known vsps in background and
	// canceled in OnNavigatedFrom().
	pg.ctx, pg.ctxCancel = context.WithCancel(context.TODO())

	pg.fetchTicketPrice()

	pg.loadPageData() // starts go routines to refresh the display which is just about to be displayed, ok?

	pg.stake.SetChecked(pg.WL.SelectedWallet.Wallet.IsAutoTicketsPurchaseActive())

	pg.setStakingButtonsState()

	pg.listenForTxNotifications()
	pg.fetchTickets()
}

// fetch ticket price only when the wallet is synced
func (pg *Page) fetchTicketPrice() {
	ticketPrice, err := pg.WL.SelectedWallet.Wallet.TicketPrice()
	if err != nil && !pg.WL.SelectedWallet.Wallet.IsSynced() {
		log.Error(err)
		pg.ticketPrice = values.String(values.StrNotAvailable)
		errModal := modal.NewErrorModal(pg.Load, values.String(values.StrWalletNotSynced), modal.DefaultClickFunc())
		pg.ParentWindow().ShowModal(errModal)
	} else {
		pg.ticketPrice = dcrutil.Amount(ticketPrice.TicketPrice).String()
	}
}

func (pg *Page) setStakingButtonsState() {
	//disable auto ticket purchase if wallet is not synced
	pg.stake.SetEnabled(pg.WL.SelectedWallet.Wallet.IsSynced() || !pg.WL.SelectedWallet.Wallet.IsWatchingOnlyWallet())
}

func (pg *Page) loadPageData() {
	go func() {
		if len(pg.WL.SelectedWallet.Wallet.KnownVSPs()) == 0 {
			// TODO: Does this page need this list?
			if pg.ctx != nil {
				pg.WL.SelectedWallet.Wallet.ReloadVSPList(pg.ctx)
			}
		}

		totalRewards, err := pg.WL.SelectedWallet.Wallet.TotalStakingRewards()
		if err != nil {
			errModal := modal.NewErrorModal(pg.Load, err.Error(), modal.DefaultClickFunc())
			pg.ParentWindow().ShowModal(errModal)
		} else {
			pg.totalRewards = dcrutil.Amount(totalRewards).String()
		}

		overview, err := pg.WL.SelectedWallet.Wallet.StakingOverview()
		if err != nil {
			errModal := modal.NewErrorModal(pg.Load, err.Error(), modal.DefaultClickFunc())
			pg.ParentWindow().ShowModal(errModal)
		} else {
			pg.ticketOverview = overview
		}

		pg.ParentWindow().Reload()
	}()
}

// Layout draws the page UI components into the provided layout context
// to be eventually drawn on screen.
// Part of the load.Page interface.
func (pg *Page) Layout(gtx C) D {
	if pg.Load.GetCurrentAppWidth() <= gtx.Dp(values.StartMobileView) {
		return pg.layoutMobile(gtx)
	}
	return pg.layoutDesktop(gtx)
}

func (pg *Page) layoutDesktop(gtx layout.Context) layout.Dimensions {
	widgets := []layout.Widget{
		func(gtx C) D {
			return components.UniformHorizontalPadding(gtx, pg.stakePriceSection)
		},
		func(gtx C) D {
			return components.UniformHorizontalPadding(gtx, pg.ticketListLayout)
		},
	}

	return layout.Inset{Top: values.MarginPadding24}.Layout(gtx, func(gtx C) D {
		return pg.Theme.List(pg.list).Layout(gtx, len(widgets), func(gtx C, i int) D {
			return widgets[i](gtx)
		})
	})
}

func (pg *Page) layoutMobile(gtx layout.Context) layout.Dimensions {
	widgets := []layout.Widget{
		pg.stakePriceSection,
		pg.ticketListLayout,
	}

	return components.UniformMobile(gtx, true, true, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: values.MarginPadding24}.Layout(gtx, func(gtx C) D {
			return pg.Theme.List(pg.list).Layout(gtx, len(widgets), func(gtx C, i int) D {
				return widgets[i](gtx)
			})
		})
	})
}

func (pg *Page) pageSections(gtx C, body layout.Widget) D {
	return layout.Inset{
		Bottom: values.MarginPadding8,
	}.Layout(gtx, func(gtx C) D {
		return pg.Theme.Card().Layout(gtx, func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.UniformInset(values.MarginPadding26).Layout(gtx, body)
		})
	})
}

// HandleUserInteractions is called just before Layout() to determine
// if any user interaction recently occurred on the page and may be
// used to update the page's UI components shortly before they are
// displayed.
// Part of the load.Page interface.
func (pg *Page) HandleUserInteractions() {
	pg.setStakingButtonsState()

	if pg.stake.Changed() {
		if pg.stake.IsChecked() {
			if pg.WL.SelectedWallet.Wallet.TicketBuyerConfigIsSet() {
				// get ticket buyer config to check if the saved wallet account is mixed
				//check if mixer is set, if yes check if allow spend from unmixed account
				//if not set, check if the saved account is mixed before opening modal
				// if it is not, open stake config modal
				tbConfig := pg.WL.SelectedWallet.Wallet.AutoTicketsBuyerConfig()
				if pg.WL.SelectedWallet.Wallet.ReadBoolConfigValueForKey(dcr.AccountMixerConfigSet, false) &&
					!pg.WL.SelectedWallet.Wallet.ReadBoolConfigValueForKey(load.SpendUnmixedFundsKey, false) &&
					(tbConfig.PurchaseAccount == pg.WL.SelectedWallet.Wallet.MixedAccountNumber()) {
					pg.startTicketBuyerPasswordModal()
				} else {
					pg.ticketBuyerSettingsModal()
				}
			} else {
				pg.ticketBuyerSettingsModal()
			}
		} else {
			pg.WL.SelectedWallet.Wallet.StopAutoTicketsPurchase()
		}
	}

	if pg.stakeSettings.Clicked() && !pg.WL.SelectedWallet.Wallet.IsWatchingOnlyWallet() {
		if pg.WL.SelectedWallet.Wallet.IsAutoTicketsPurchaseActive() {
			errModal := modal.NewErrorModal(pg.Load, values.String(values.StrAutoTicketWarn), modal.DefaultClickFunc())
			pg.ParentWindow().ShowModal(errModal)
			return
		}

		ticketBuyerModal := newTicketBuyerModal(pg.Load).
			OnSettingsSaved(func() {
				infoModal := modal.NewSuccessModal(pg.Load, values.String(values.StrTicketSettingSaved), modal.DefaultClickFunc())
				pg.ParentWindow().ShowModal(infoModal)
			}).
			OnCancel(func() {
				pg.stake.SetChecked(false)
			})
		pg.ParentWindow().ShowModal(ticketBuyerModal)
	}

	secs, _ := pg.WL.SelectedWallet.Wallet.NextTicketPriceRemaining()
	if secs <= 0 {
		pg.fetchTicketPrice()
	}

	if pg.WL.SelectedWallet.Wallet.IsSynced() {
		pg.fetchTicketPrice()
	}

	if clicked, selectedItem := pg.ticketsList.ItemClicked(); clicked {
		ticketTx := pg.tickets[selectedItem].transaction
		pg.ParentNavigator().Display(tpage.NewTransactionDetailsPage(pg.Load, ticketTx, true))

		// Check if this ticket is fully registered with a VSP
		// and log any discrepancies.
		// NOTE: Wallet needs to be unlocked to get the ticket status
		// from the vsp. Otherwise, only the wallet-stored info will
		// be retrieved. This is fine because we're only just logging
		// but where it is necessary to display vsp-stored info, the
		// wallet passphrase should be requested and used to unlock
		// the wallet before calling this method.
		// TODO: Use log.Errorf and log.Warnf instead of fmt.Printf.
		ticketInfo, err := pg.WL.SelectedWallet.Wallet.VSPTicketInfo(ticketTx.Hash)
		if err != nil {
			log.Errorf("VSPTicketInfo error: %v\n", err)
		} else {
			if ticketInfo.FeeTxStatus != dcr.VSPFeeProcessConfirmed {
				log.Errorf("[WARN] Ticket %s has unconfirmed fee tx %s with status %q, vsp %s \n",
					ticketTx.Hash, ticketInfo.FeeTxHash, ticketInfo.FeeTxStatus.String(), ticketInfo.VSP)
			}
			if ticketInfo.ConfirmedByVSP == nil || !*ticketInfo.ConfirmedByVSP {
				log.Errorf("[WARN] Ticket %s is not confirmed by VSP %s. Fee tx %s, status %q \n",
					ticketTx.Hash, ticketInfo.VSP, ticketInfo.FeeTxHash, ticketInfo.FeeTxStatus.String())
			}
		}
	}

	if pg.infoButton.Button.Clicked() {
		backupNowOrLaterModal := modal.NewCustomModal(pg.Load).
			Title(values.String(values.StrStatistics)).
			SetCancelable(true).
			UseCustomWidget(func(gtx C) D {
				return pg.stakingRecordStatistics(gtx)
			}).
			SetPositiveButtonText(values.String(values.StrGotIt))
		pg.ParentWindow().ShowModal(backupNowOrLaterModal)
	}
}

func (pg *Page) ticketBuyerSettingsModal() {
	ticketBuyerModal := newTicketBuyerModal(pg.Load).
		OnCancel(func() {
			pg.stake.SetChecked(false)
		}).
		OnSettingsSaved(func() {
			pg.startTicketBuyerPasswordModal()
			infoModal := modal.NewSuccessModal(pg.Load, values.String(values.StrTicketSettingSaved), modal.DefaultClickFunc())
			pg.ParentWindow().ShowModal(infoModal)
		})
	pg.ParentWindow().ShowModal(ticketBuyerModal)
}

func (pg *Page) startTicketBuyerPasswordModal() {
	tbConfig := pg.WL.SelectedWallet.Wallet.AutoTicketsBuyerConfig()
	balToMaintain := dcr.AmountCoin(tbConfig.BalanceToMaintain)
	name, err := pg.WL.SelectedWallet.Wallet.AccountNameRaw(uint32(tbConfig.PurchaseAccount))
	if err != nil {
		errModal := modal.NewErrorModal(pg.Load, values.StringF(values.StrTicketError, err), modal.DefaultClickFunc())
		pg.ParentWindow().ShowModal(errModal)
		return
	}

	walletPasswordModal := modal.NewCreatePasswordModal(pg.Load).
		EnableName(false).
		EnableConfirmPassword(false).
		Title(values.String(values.StrConfirmPurchase)).
		SetCancelable(false).
		UseCustomWidget(func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(pg.Theme.Label(values.TextSize14, values.StringF(values.StrWalletToPurchaseFrom, pg.WL.SelectedWallet.Wallet.Name)).Layout),
				layout.Rigid(pg.Theme.Label(values.TextSize14, values.StringF(values.StrSelectedAccount, name)).Layout),
				layout.Rigid(pg.Theme.Label(values.TextSize14, values.StringF(values.StrBalToMaintainValue, balToMaintain)).Layout), layout.Rigid(func(gtx C) D {
					label := pg.Theme.Label(values.TextSize14, fmt.Sprintf("VSP: %s", tbConfig.VspHost))
					return layout.Inset{Bottom: values.MarginPadding12}.Layout(gtx, label.Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return cryptomaterial.LinearLayout{
						Width:      cryptomaterial.MatchParent,
						Height:     cryptomaterial.WrapContent,
						Background: pg.Theme.Color.LightBlue,
						Padding: layout.Inset{
							Top:    values.MarginPadding12,
							Bottom: values.MarginPadding12,
						},
						Border:    cryptomaterial.Border{Radius: cryptomaterial.Radius(8)},
						Direction: layout.Center,
						Alignment: layout.Middle,
					}.Layout2(gtx, func(gtx C) D {
						return layout.Inset{Bottom: values.MarginPadding4}.Layout(gtx, func(gtx C) D {
							msg := values.String(values.StrAutoTicketInfo)
							txt := pg.Theme.Label(values.TextSize14, msg)
							txt.Alignment = text.Middle
							txt.Color = pg.Theme.Color.GrayText3
							if pg.WL.SelectedWallet.Wallet.ReadBoolConfigValueForKey(load.DarkModeConfigKey, false) {
								txt.Color = pg.Theme.Color.Gray3
							}
							return txt.Layout(gtx)
						})
					})
				}),
			)
		}).
		SetNegativeButtonCallback(func() { pg.stake.SetChecked(false) }).
		SetPositiveButtonCallback(func(_, password string, pm *modal.CreatePasswordModal) bool {
			if !pg.WL.SelectedWallet.Wallet.IsConnectedToDecredNetwork() {
				pm.SetError(values.String(values.StrNotConnected))
				pm.SetLoading(false)
				pg.stake.SetChecked(false)
				return false
			}

			err := pg.WL.SelectedWallet.Wallet.StartTicketBuyer([]byte(password))
			if err != nil {
				pm.SetError(err.Error())
				pm.SetLoading(false)
				return false
			}

			pg.stake.SetChecked(pg.WL.SelectedWallet.Wallet.IsAutoTicketsPurchaseActive())
			pg.ParentWindow().Reload()

			pm.Dismiss()

			return false
		})
	pg.ParentWindow().ShowModal(walletPasswordModal)
}

// OnNavigatedFrom is called when the page is about to be removed from
// the displayed window. This method should ideally be used to disable
// features that are irrelevant when the page is NOT displayed.
// NOTE: The page may be re-displayed on the app's window, in which case
// OnNavigatedTo() will be called again. This method should not destroy UI
// components unless they'll be recreated in the OnNavigatedTo() method.
// Part of the load.Page interface.
func (pg *Page) OnNavigatedFrom() {
	pg.ctxCancel()
}