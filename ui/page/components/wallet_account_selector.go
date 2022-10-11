package components

import (
	"context"
	"errors"

	"gioui.org/io/event"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"

	"github.com/decred/dcrd/dcrutil/v4"
	"gitlab.com/raedah/cryptopower/app"
	"gitlab.com/raedah/cryptopower/libwallet/wallets/dcr"
	"gitlab.com/raedah/cryptopower/listeners"
	"gitlab.com/raedah/cryptopower/ui/cryptomaterial"
	"gitlab.com/raedah/cryptopower/ui/load"
	"gitlab.com/raedah/cryptopower/ui/renderers"
	"gitlab.com/raedah/cryptopower/ui/values"
)

const WalletAndAccountSelectorID = "WalletAndAccountSelector"

type WalletAndAccountSelector struct {
	*load.Load
	*listeners.TxAndBlockNotificationListener

	selectedWallet  *dcr.Wallet
	selectedAccount *dcr.Account
	accountCallback func(*dcr.Account)
	walletCallback  func(*dcr.Wallet)
	accountIsValid  func(*dcr.Account) bool
	accountSelector bool

	openSelectorDialog *cryptomaterial.Clickable
	selectorModal      *SelectorModal
	infoActionText     string

	dialogTitle  string
	totalBalance string
	changed      bool
}

// NewWalletAndAccountSelector creates a wallet selector component.
// It opens a modal to select a desired wallet or a desired account.
func NewWalletAndAccountSelector(l *load.Load) *WalletAndAccountSelector {
	return &WalletAndAccountSelector{
		Load:               l,
		openSelectorDialog: l.Theme.NewClickable(true),
		accountIsValid:     func(*dcr.Account) bool { return false },
		selectedWallet:     l.WL.SelectedWallet.Wallet, // Set the default wallet to wallet loaded by cryptopower.
		accountSelector:    false,
	}
}

// SelectedAccount returns the currently selected account.
func (ws *WalletAndAccountSelector) SelectedAccount() *dcr.Account {
	return ws.selectedAccount
}

// AccountValidator validates an account according to the rules defined to determine a valid a account.
func (ws *WalletAndAccountSelector) AccountValidator(accountIsValid func(*dcr.Account) bool) *WalletAndAccountSelector {
	ws.accountIsValid = accountIsValid
	return ws
}

// SetActionInfoText sets the text that is shown when the info action icon of the selector
// modal is is clicked. The {text} is rendered using a html renderer. So HTML text can be passed in.
func (ws *WalletAndAccountSelector) SetActionInfoText(text string) *WalletAndAccountSelector {
	ws.infoActionText = text
	return ws
}

// SelectFirstValidAccount transforms this widget into an Account selector and selects the first valid account from the
// the wallet passed to this method.
func (ws *WalletAndAccountSelector) SelectFirstValidAccount(wallet *dcr.Wallet) error {
	if !ws.accountSelector {
		ws.accountSelector = true
	}
	ws.SetSelectedWallet(wallet)

	accountsResult, err := wallet.GetAccountsRaw()
	if err != nil {
		return err
	}

	accounts := accountsResult.Acc
	for _, account := range accounts {
		if ws.accountIsValid(account) {
			ws.SetSelectedAccount(account)
			if ws.accountCallback != nil {
				ws.accountCallback(account)
			}
			return nil
		}
	}

	return errors.New(values.String(values.StrNoValidAccountFound))
}

func (ws *WalletAndAccountSelector) SetSelectedAccount(account *dcr.Account) {
	ws.selectedAccount = account
	ws.totalBalance = dcrutil.Amount(account.TotalBalance).String()
}

func (ws *WalletAndAccountSelector) Clickable() *cryptomaterial.Clickable {
	return ws.openSelectorDialog
}

func (ws *WalletAndAccountSelector) Title(title string) *WalletAndAccountSelector {
	ws.dialogTitle = title
	return ws
}

func (ws *WalletAndAccountSelector) WalletSelected(callback func(*dcr.Wallet)) *WalletAndAccountSelector {
	ws.walletCallback = callback
	return ws
}

func (ws *WalletAndAccountSelector) AccountSelected(callback func(*dcr.Account)) *WalletAndAccountSelector {
	ws.accountCallback = callback
	return ws
}

func (ws *WalletAndAccountSelector) Changed() bool {
	changed := ws.changed
	ws.changed = false
	return changed
}

func (ws *WalletAndAccountSelector) Handle(window app.WindowNavigator) {
	for ws.openSelectorDialog.Clicked() {
		ws.selectorModal = newSelectorModal(ws.Load, ws).
			title(ws.dialogTitle).
			accountValidator(ws.accountIsValid).
			walletSelected(func(wallet *dcr.Wallet) {
				ws.changed = true
				ws.SetSelectedWallet(wallet)
				if ws.walletCallback != nil {
					ws.walletCallback(wallet)
				}
			}).
			accountSelected(func(account *dcr.Account) {
				if ws.selectedAccount.Number != account.Number {
					ws.changed = true
				}
				ws.SetSelectedAccount(account)
				if ws.accountCallback != nil {
					ws.accountCallback(account)
				}
			}).
			onModalExit(func() {
				ws.selectorModal = nil
			})
		window.ShowModal(ws.selectorModal)
	}
}

func (ws *WalletAndAccountSelector) SetSelectedWallet(wallet *dcr.Wallet) {
	ws.selectedWallet = wallet
}

func (ws *WalletAndAccountSelector) SelectedWallet() *dcr.Wallet {
	return ws.selectedWallet
}

func (ws *WalletAndAccountSelector) Layout(window app.WindowNavigator, gtx C) D {
	ws.Handle(window)

	return cryptomaterial.LinearLayout{
		Width:   cryptomaterial.MatchParent,
		Height:  cryptomaterial.WrapContent,
		Padding: layout.UniformInset(values.MarginPadding12),
		Border: cryptomaterial.Border{
			Width:  values.MarginPadding2,
			Color:  ws.Theme.Color.Gray2,
			Radius: cryptomaterial.Radius(8),
		},
		Clickable: ws.Clickable(),
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			walletIcon := ws.Theme.Icons.DecredLogo
			if ws.accountSelector {
				walletIcon = ws.Theme.Icons.AccountIcon
			}
			inset := layout.Inset{
				Right: values.MarginPadding8,
			}
			return inset.Layout(gtx, func(gtx C) D {
				return walletIcon.Layout24dp(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			if ws.accountSelector {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
					layout.Rigid(ws.Theme.Body1(ws.SelectedAccount().Name).Layout),
					layout.Rigid(func(gtx C) D {
						walName := ws.Theme.Label(values.TextSize12, ws.SelectedWallet().Name)
						walName.Color = ws.Theme.Color.GrayText2
						card := ws.Theme.Card()
						card.Radius = cryptomaterial.Radius(4)
						card.Color = ws.Theme.Color.Gray4
						return layout.Inset{
							Left: values.MarginPadding4,
						}.Layout(gtx, func(gtx C) D {
							return card.Layout(gtx, func(gtx C) D {
								return layout.UniformInset(values.MarginPadding0).Layout(gtx, walName.Layout)
							})
						})

					}),
				)
			}
			return ws.Theme.Body1(ws.SelectedWallet().Name).Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						if ws.accountSelector {
							return ws.Theme.Body1(ws.totalBalance).Layout(gtx)
						}
						totalBal, _ := walletBalance(ws.SelectedWallet())
						return ws.Theme.Body1(dcrutil.Amount(totalBal).String()).Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						inset := layout.Inset{
							Left: values.MarginPadding15,
						}
						return inset.Layout(gtx, func(gtx C) D {
							ic := cryptomaterial.NewIcon(ws.Theme.Icons.DropDownIcon)
							ic.Color = ws.Theme.Color.Gray1
							return ic.Layout(gtx, values.MarginPadding20)
						})
					}),
				)
			})
		}),
	)
}

func (ws *WalletAndAccountSelector) ListenForTxNotifications(ctx context.Context, window app.WindowNavigator) {
	if ws.TxAndBlockNotificationListener != nil {
		return
	}
	ws.TxAndBlockNotificationListener = listeners.NewTxAndBlockNotificationListener()
	err := ws.WL.SelectedWallet.Wallet.AddTxAndBlockNotificationListener(ws.TxAndBlockNotificationListener, true, WalletAndAccountSelectorID)
	if err != nil {
		log.Errorf("Error adding tx and block notification listener: %v", err)
		return
	}

	go func() {
		for {
			select {
			case n := <-ws.TxAndBlockNotifChan:
				switch n.Type {
				case listeners.BlockAttached:
					// refresh wallet and account balance on every new block
					// only if sync is completed.
					if ws.WL.SelectedWallet.Wallet.IsSynced() {
						if ws.selectorModal != nil {
							if ws.accountSelector {
								ws.selectorModal.setupAccounts(ws.selectedWallet)
							} else {
								ws.selectorModal.setupWallet()
							}
							window.Reload()
						}

					}
				case listeners.NewTransaction:
					// refresh wallets/Accounts list when new transaction is received
					if ws.selectorModal != nil {
						if ws.accountSelector {
							ws.selectorModal.setupAccounts(ws.selectedWallet)
						} else {
							ws.selectorModal.setupWallet()
						}
						window.Reload()
					}

				}
			case <-ctx.Done():
				ws.WL.SelectedWallet.Wallet.RemoveTxAndBlockNotificationListener(WalletAndAccountSelectorID)
				close(ws.TxAndBlockNotifChan)
				ws.TxAndBlockNotificationListener = nil
				return
			}
		}
	}()
}

type SelectorModal struct {
	*load.Load
	*cryptomaterial.Modal

	walletIsValid   func(*dcr.Wallet) bool
	accountIsValid  func(*dcr.Account) bool
	walletCallback  func(*dcr.Wallet)
	accountCallback func(*dcr.Account)
	onExit          func()

	walletsList layout.List

	currentSelectedWallet *dcr.Wallet
	selectorItems         []*SelectorItem // A SelectorItem can either be a wallet or account
	eventQueue            event.Queue

	dialogTitle string

	isCancelable   bool
	walletSelector *WalletAndAccountSelector
	infoButton     cryptomaterial.IconButton
	infoModalOpen  bool
	infoBackdrop   *widget.Clickable
}

// SelectorItem models a wallet or an account a long with it's clickable.
type SelectorItem struct {
	item      interface{} // Item can either be a wallet or an account.
	clickable *cryptomaterial.Clickable
}

func newSelectorModal(l *load.Load, ws *WalletAndAccountSelector) *SelectorModal {
	sm := &SelectorModal{
		Load:                  l,
		Modal:                 l.Theme.ModalFloatTitle("SelectorModal"),
		walletsList:           layout.List{Axis: layout.Vertical},
		currentSelectedWallet: ws.selectedWallet,
		isCancelable:          true,
		walletSelector:        ws,
		infoBackdrop:          new(widget.Clickable),
	}

	sm.infoButton = l.Theme.IconButton(l.Theme.Icons.ActionInfo)
	sm.infoButton.Size = values.MarginPadding14
	sm.infoButton.Inset = layout.UniformInset(values.MarginPadding4)

	sm.Modal.ShowScrollbar(true)
	return sm
}

func (sm *SelectorModal) OnResume() {
	if sm.walletSelector.accountSelector {
		sm.setupAccounts(sm.currentSelectedWallet)
		return
	}
	sm.setupWallet()
}

func (sm *SelectorModal) setupWallet() {
	selectorItems := make([]*SelectorItem, 0)
	wallets := sm.WL.SortedWalletList()
	for _, wal := range wallets {
		if !sm.WL.SelectedWallet.Wallet.IsWatchingOnlyWallet() {
			selectorItems = append(selectorItems, &SelectorItem{
				item:      wal,
				clickable: sm.Theme.NewClickable(true),
			})
		}
	}
	sm.selectorItems = selectorItems
}

func (sm *SelectorModal) setupAccounts(wal *dcr.Wallet) {
	selectorItems := make([]*SelectorItem, 0)
	if !wal.IsWatchingOnlyWallet() {
		accountsResult, err := wal.GetAccountsRaw()
		if err != nil {
			log.Errorf("Error getting accounts:", err)
			return
		}

		for _, account := range accountsResult.Acc {
			if sm.accountIsValid(account) {
				selectorItems = append(selectorItems, &SelectorItem{
					item:      account,
					clickable: sm.Theme.NewClickable(true),
				})
			}
		}
	}
	sm.selectorItems = selectorItems
}

func (sm *SelectorModal) SetCancelable(min bool) *SelectorModal {
	sm.isCancelable = min
	return sm
}

func (sm *SelectorModal) accountValidator(accountIsValid func(*dcr.Account) bool) *SelectorModal {
	sm.accountIsValid = accountIsValid
	return sm
}

func (sm *SelectorModal) Handle() {
	if sm.eventQueue != nil {
		for _, selectorItem := range sm.selectorItems {
			for selectorItem.clickable.Clicked() {
				switch item := selectorItem.item.(type) {
				case *dcr.Account:
					if sm.accountCallback != nil {
						sm.accountCallback(item)
					}
				case *dcr.Wallet:
					if sm.walletCallback != nil {
						sm.walletCallback(item)
					}
				}
				sm.onExit()
				sm.Dismiss()
			}
		}

		if sm.infoBackdrop.Clicked() {
			sm.infoModalOpen = false
		}

		if sm.infoButton.IconButtonStyle.Button.Clicked() {
			sm.infoModalOpen = !sm.infoModalOpen
		}
	}

	if sm.Modal.BackdropClicked(sm.isCancelable) {
		sm.onExit()
		sm.Dismiss()
	}
}

func (sm *SelectorModal) title(title string) *SelectorModal {
	sm.dialogTitle = title
	return sm
}

func (sm *SelectorModal) walletSelected(callback func(*dcr.Wallet)) *SelectorModal {
	sm.walletCallback = callback
	return sm
}

func (sm *SelectorModal) accountSelected(callback func(*dcr.Account)) *SelectorModal {
	sm.accountCallback = callback
	return sm
}

func (sm *SelectorModal) Layout(gtx C) D {
	sm.eventQueue = gtx
	sm.infoBackdropLayout(gtx)

	w := []layout.Widget{
		func(gtx C) D {
			title := sm.Theme.H6(sm.dialogTitle)
			title.Color = sm.Theme.Color.Text
			title.Font.Weight = text.SemiBold
			return layout.Inset{
				Top: values.MarginPaddingMinus15,
			}.Layout(gtx, func(gtx C) D {
				return title.Layout(gtx)
			})
		},
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if sm.walletSelector.accountSelector {
						inset := layout.Inset{
							Top: values.MarginPadding0,
						}
						return inset.Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									inset := layout.UniformInset(values.MarginPadding4)
									return inset.Layout(gtx, func(gtx C) D {
										walName := sm.Theme.Label(values.TextSize14, sm.currentSelectedWallet.Name)
										walName.Color = sm.Theme.Color.GrayText2
										return walName.Layout(gtx)
									})
								}),

								layout.Rigid(func(gtx C) D {
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
										layout.Rigid(func(gtx C) D {
											if sm.infoModalOpen {
												m := op.Record(gtx.Ops)
												layout.Inset{Top: values.MarginPadding30}.Layout(gtx, func(gtx C) D {
													card := sm.Theme.Card()
													card.Color = sm.Theme.Color.Surface
													return card.Layout(gtx, func(gtx C) D {
														return layout.UniformInset(values.MarginPadding12).Layout(gtx, renderers.RenderHTML(sm.walletSelector.infoActionText, sm.Theme).Layout)
													})
												})
												op.Defer(gtx.Ops, m.Stop())
											}

											return D{}
										}),
										layout.Rigid(func(gtx C) D {
											if sm.walletSelector.infoActionText != "" {
												return sm.infoButton.Layout(gtx)
											}
											return D{}
										}),
									)
								}),
							)
						})
					}
					return D{}
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Stack{Alignment: layout.NW}.Layout(gtx,
						layout.Expanded(func(gtx C) D {
							return sm.walletsList.Layout(gtx, len(sm.selectorItems), func(gtx C, aindex int) D {
								return sm.modalListItemLayout(gtx, sm.selectorItems[aindex])
							})
						}),
					)

				}),
			)
		},
	}

	return sm.Modal.Layout(gtx, w)
}

// infoBackdropLayout draws background overlay when the confirmation modal action button is clicked.
func (sm *SelectorModal) infoBackdropLayout(gtx C) {
	if sm.infoModalOpen {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		m := op.Record(gtx.Ops)
		sm.infoBackdrop.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.Button.Add(gtx.Ops)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		})
		op.Defer(gtx.Ops, m.Stop())
	}
}

func walletBalance(wal *dcr.Wallet) (totalBalance, spendableBalance int64) {
	accountsResult, _ := wal.GetAccountsRaw()
	var tBal, sBal int64
	for _, account := range accountsResult.Acc {
		tBal += account.TotalBalance
		sBal += account.Balance.Spendable
	}
	return tBal, sBal
}

func (sm *SelectorModal) modalListItemLayout(gtx C, selectorItem *SelectorItem) D {
	accountIcon := sm.Theme.Icons.AccountIcon

	return cryptomaterial.LinearLayout{
		Width:     cryptomaterial.MatchParent,
		Height:    cryptomaterial.WrapContent,
		Margin:    layout.Inset{Bottom: values.MarginPadding4},
		Padding:   layout.Inset{Top: values.MarginPadding8, Bottom: values.MarginPadding8},
		Clickable: selectorItem.clickable,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(0.1, func(gtx C) D {
			return layout.Inset{
				Right: values.MarginPadding18,
			}.Layout(gtx, accountIcon.Layout16dp)
		}),
		layout.Flexed(0.8, func(gtx C) D {
			var name, totalBal, spendableBal string
			switch t := selectorItem.item.(type) {
			case *dcr.Account:
				totalBal = dcrutil.Amount(t.TotalBalance).String()
				spendableBal = dcrutil.Amount(t.Balance.Spendable).String()
				name = t.Name
			case *dcr.Wallet:
				tb, sb := walletBalance(t)
				totalBal = dcrutil.Amount(tb).String()
				spendableBal = dcrutil.Amount(sb).String()
				name = t.Name
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					acct := sm.Theme.Label(values.TextSize18, name)
					acct.Color = sm.Theme.Color.Text
					acct.Font.Weight = text.Normal
					return EndToEndRow(gtx, acct.Layout, func(gtx C) D {
						return LayoutBalance(gtx, sm.Load, totalBal)
					})
				}),
				layout.Rigid(func(gtx C) D {
					spendableText := sm.Theme.Label(values.TextSize14, values.String(values.StrLabelSpendable))
					spendableText.Color = sm.Theme.Color.GrayText2
					spendableLabel := sm.Theme.Label(values.TextSize14, spendableBal)
					spendableLabel.Color = sm.Theme.Color.GrayText2
					return EndToEndRow(gtx, spendableText.Layout, spendableLabel.Layout)
				}),
			)
		}),

		layout.Flexed(0.1, func(gtx C) D {
			inset := layout.Inset{
				Top:  values.MarginPadding10,
				Left: values.MarginPadding10,
			}
			sections := func(gtx C) D {
				return layout.E.Layout(gtx, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						ic := cryptomaterial.NewIcon(sm.Theme.Icons.NavigationCheck)
						ic.Color = sm.Theme.Color.Gray1
						return ic.Layout(gtx, values.MarginPadding20)
					})
				})
			}
			switch t := selectorItem.item.(type) {
			case *dcr.Account:
				if t.Number == sm.walletSelector.selectedAccount.Number {
					return sections(gtx)
				}
			case *dcr.Wallet:
				if t.ID == sm.currentSelectedWallet.ID {
					return sections(gtx)
				}
			}
			return D{}
		}),
	)
}

func (sm *SelectorModal) onModalExit(exit func()) *SelectorModal {
	sm.onExit = exit
	return sm
}

func (sm *SelectorModal) OnDismiss() {}
