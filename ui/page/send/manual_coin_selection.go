package send

import (
	"code.cryptopower.dev/group/cryptopower/app"
	sharedW "code.cryptopower.dev/group/cryptopower/libwallet/assets/wallet"
	"code.cryptopower.dev/group/cryptopower/ui/cryptomaterial"
	"code.cryptopower.dev/group/cryptopower/ui/load"
	"code.cryptopower.dev/group/cryptopower/ui/page/components"
	"code.cryptopower.dev/group/cryptopower/ui/values"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
)

const ManualCoinSelectionPageID = "backup_success"

type AccountsUTXOInfo struct {
	AccountName string
	Details     []*sharedW.UnspentOutput
}

type ManualCoinSelectionPage struct {
	*load.Load
	// GenericPageModal defines methods such as ID() and OnAttachedToNavigator()
	// that helps this Page satisfy the app.Page interface. It also defines
	// helper methods for accessing the PageNavigator that displayed this page
	// and the root WindowNavigator.
	*app.GenericPageModal

	actionButton cryptomaterial.Button
	clearButton  cryptomaterial.Button

	selectedUTXOs cryptomaterial.Label
	txSize        cryptomaterial.Label
	totalAmount   cryptomaterial.Label

	accountsUTXOs []*AccountsUTXOInfo
	UTXOList      *cryptomaterial.ClickableList

	listContainer *widget.List
}

func NewManualCoinSelectionPage(l *load.Load) *ManualCoinSelectionPage {
	pg := &ManualCoinSelectionPage{
		Load:             l,
		GenericPageModal: app.NewGenericPageModal(ManualCoinSelectionPageID),

		actionButton: l.Theme.OutlineButton(values.String(values.StrDone)),
		clearButton:  l.Theme.Button(values.String(values.StrClearSelection)),

		listContainer: &widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		UTXOList: l.Theme.NewClickableList(layout.Vertical),
	}
	pg.actionButton.Font.Weight = text.Medium
	pg.clearButton.Font.Weight = text.Medium
	pg.clearButton.Color = l.Theme.Color.Danger

	pg.selectedUTXOs = pg.Theme.Label(values.TextSize16, "4")
	pg.txSize = pg.Theme.Label(values.TextSize16, "1200 bytes")
	pg.totalAmount = pg.Theme.Label(values.TextSize16, "3 BTC")

	pg.selectedUTXOs.Font.Weight = text.SemiBold
	pg.txSize.Font.Weight = text.SemiBold
	pg.totalAmount.Font.Weight = text.SemiBold

	return pg
}

// OnNavigatedTo is called when the page is about to be displayed and
// may be used to initialize page features that are only relevant when
// the page is displayed.
// Part of the load.Page interface.
func (pg *ManualCoinSelectionPage) OnNavigatedTo() {}

// HandleUserInteractions is called just before Layout() to determine
// if any user interaction recently occurred on the page and may be
// used to update the page's UI components shortly before they are
// displayed.
// Part of the load.Page interface.
func (pg *ManualCoinSelectionPage) HandleUserInteractions() {
	if pg.actionButton.Clicked() {
	}
}

// OnNavigatedFrom is called when the page is about to be removed from
// the displayed window. This method should ideally be used to disable
// features that are irrelevant when the page is NOT displayed.
// NOTE: The page may be re-displayed on the app's window, in which case
// OnNavigatedTo() will be called again. This method should not destroy UI
// components unless they'll be recreated in the OnNavigatedTo() method.
// Part of the load.Page interface.
func (pg *ManualCoinSelectionPage) OnNavigatedFrom() {}

// Layout draws the page UI components into the provided layout context
// to be eventually drawn on screen.
// Part of the load.Page interface.
func (pg *ManualCoinSelectionPage) Layout(gtx C) D {
	return components.UniformPadding(gtx, func(gtx C) D {
		return cryptomaterial.LinearLayout{
			Width:       cryptomaterial.MatchParent,
			Height:      cryptomaterial.MatchParent,
			Orientation: layout.Vertical,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx C) D {
				return cryptomaterial.LinearLayout{
					Width:       cryptomaterial.MatchParent,
					Height:      cryptomaterial.MatchParent,
					Orientation: layout.Vertical,
					Alignment:   layout.Middle,
					Direction:   layout.Center,
				}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						successIcon := cryptomaterial.NewIcon(pg.Theme.Icons.ActionCheckCircle)
						return successIcon.Layout(gtx, values.MarginPadding64)
					}),
					layout.Rigid(func(gtx C) D {
						label := pg.Theme.Label(values.TextSize24, values.String(values.StrSeedPhraseVerified))
						label.Color = pg.Theme.Color.DeepBlue

						return layout.Inset{Top: values.MarginPadding24}.Layout(gtx, label.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						label := pg.Theme.Label(values.TextSize16, values.String(values.StrSureToSafeStoreSeed))
						label.Color = pg.Theme.Color.GrayText1

						return layout.Inset{Top: values.MarginPadding16}.Layout(gtx, label.Layout)
					}),
				)
			}),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X

				return layout.Inset{Top: values.MarginPadding16}.Layout(gtx, pg.actionButton.Layout)
			}),
		)
	})
}

func (pg *ManualCoinSelectionPage) summarySection(gtx C) D {
	return layout.Inset{}.Layout(gtx, func(gtx C) D {
		return pg.Theme.Card().Layout(gtx, func(gtx C) D {
			topContainer := layout.UniformInset(values.MarginPadding15)
			return topContainer.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X // use maximum width
				textLabel := pg.Theme.Label(values.TextSize16, values.String(values.StrSummary))
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(textLabel.Layout),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Flexed(0.3, func(gtx C) D {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(pg.Theme.Label(values.TextSize16, values.String(values.StrSelectedUTXO)+":").Layout),
									layout.Flexed(1, func(gtx C) D {
										return layout.W.Layout(gtx, pg.selectedUTXOs.Layout)
									}),
								)
							}),
							layout.Flexed(0.3, func(gtx C) D {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(pg.Theme.Label(values.TextSize16, values.String(values.StrTxSize)+":").Layout),
									layout.Flexed(1, func(gtx C) D {
										return layout.W.Layout(gtx, pg.selectedUTXOs.Layout)
									}),
								)
							}),
							layout.Flexed(0.3, func(gtx C) D {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(pg.Theme.Label(values.TextSize16, values.String(values.StrTotalAmount)+":").Layout),
									layout.Flexed(1, func(gtx C) D {
										return layout.W.Layout(gtx, pg.selectedUTXOs.Layout)
									}),
								)
							}),
						)
					}),
				)
			})
		})
	})
}

func (pg *ManualCoinSelectionPage) accountListSection(gtx C) D {
	return layout.Inset{}.Layout(gtx, func(gtx C) D {
		return pg.Theme.Card().Layout(gtx, func(gtx C) D {
			topContainer := layout.UniformInset(values.MarginPadding15)
			return topContainer.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X // use maximum width
				textLabel := pg.Theme.Label(values.TextSize16, values.String(values.StrAccountList))

				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(textLabel.Layout),
							layout.Flexed(1, func(gtx C) D {
								return layout.E.Layout(gtx, pg.clearButton.Layout)
							}),
						)
					}),
					layout.Rigid(func(gtx C) D {
						return pg.Theme.List(pg.listContainer).Layout(gtx, len(pg.accountsUTXOs), func(gtx C, i int) D {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X // use maximum width

							return layout.Inset{
								Bottom: values.MarginPadding4,
								Top:    values.MarginPadding4,
							}.Layout(gtx, func(gtx C) D {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										t := pg.Theme.Label(values.TextSize16, pg.accountsUTXOs[i].AccountName)
										return t.Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return pg.accountListItemsSection(gtx, pg.accountsUTXOs[i].Details)
									}),
								)
							})
						})
					}),
				)
			})
		})
	})
}

func (pg *ManualCoinSelectionPage) accountListItemsSection(gtx C, utxos []*sharedW.UnspentOutput) D {
	return layout.Inset{Right: values.MarginPadding2}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				amountLabel := pg.Theme.Label(values.TextSize16, values.String(values.StrAmount))
				addresslabel := pg.Theme.Label(values.TextSize16, values.String(values.StrAddress))
				confirmationsLabel := pg.Theme.Label(values.TextSize16, values.String(values.StrConfirmations))
				dateLabel := pg.Theme.Label(values.TextSize16, values.String(values.StrDateCreated))
			}),
			layout.Rigid(func(gtx C) D {
				return pg.UTXOList.Layout(gtx, len(utxos), func(gtx C, index int) D {
					row := components.TransactionRow{
						Transaction: utxos[index],
						Index:       index,
					}

					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return components.LayoutTransactionRow(gtx, pg.Load, row)
						}),
						layout.Rigid(func(gtx C) D {
							// No divider for last row
							if row.Index == len(utxos)-1 {
								return layout.Dimensions{}
							}

							gtx.Constraints.Min.X = gtx.Constraints.Max.X
							separator := pg.Theme.Separator()
							return layout.E.Layout(gtx, func(gtx C) D {
								// Show bottom divider for all rows except last
								return layout.Inset{Left: values.MarginPadding56}.Layout(gtx, separator.Layout)
							})
						}),
					)
				})
			}),
		)
	})
}

func (pg *ManualCoinSelectionPage) accountListItemsSection(gtx C, component1, component2, component3, component4 layout.Widget) D {
	return cryptomaterial.LinearLayout{
		Orientation: layout.Horizontal,
		Width:       cryptomaterial.MatchParent,
		Height:      gtx.Dp(values.MarginPadding56),
		Alignment:   layout.Middle,
		Padding:     layout.Inset{Left: values.MarginPadding16, Right: values.MarginPadding16},
	}.Layout(gtx,
		layout.Rigid(component1.Layout),
		layout.Rigid(component2.Layout),
		layout.Rigid(component3.Layout),
		layout.Rigid(component4.Layout),
	)
}
