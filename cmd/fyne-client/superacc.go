package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
	"image/color"
	"log"
	"strings"
	"time"

	superaccpb "rubr/proto/superacc"
)

//************************
//* Page with all groups *
//************************

var GroupName string

type GroupEntry struct {
	NameEntry        *widget.Entry
	DescriptionEntry *widget.Entry
	CommentEntry     *widget.Entry
	EvaluationEntry  *widget.Entry
	Container        *fyne.Container
}

func СreateGroupListPage(state *AppState) fyne.CanvasObject {
	w := state.window
	headerTextColor := color.White

	logoText := canvas.NewText("ВШЭ", headerTextColor)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText("Список групп", headerTextColor)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		container.NewCenter(headerTitle),
	)

	backButton := widget.NewButton("Назад", func() {
		dialog.ShowConfirm(
			"Подтверждение",
			"Выйти из приложения?",
			func(ok bool) {
				if ok {
					state.currentPage = "greeting"
					w.SetContent(createContent(state))
					return
				}
			},
			w,
		)
	})
	backButtonContainer := container.NewHBox(layout.NewSpacer(), backButton)

	groupInfoListContainer := container.NewVBox()

	columnHeaders := container.New(layout.NewGridLayoutWithColumns(4),
		container.NewPadded(widget.NewLabelWithStyle("Название группы", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Описание", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		container.NewPadded(widget.NewLabelWithStyle("Дисциплины", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)

	var groups []*GroupEntry

	addCriterionEntry := func(group *superaccpb.Group) {
		nameEntry := widget.NewEntry()
		nameEntry.SetText(group.Name)
		nameEntry.Disable()
		nameEntryContainer := container.NewMax(nameEntry)
		nameEntryContainer.Resize(fyne.NewSize(250, 60))

		descriptionEntry := widget.NewEntry()
		descriptionEntry.SetText(group.Description)
		descriptionEntry.Disable()
		descriptionEntryContainer := container.NewMax(descriptionEntry)
		descriptionEntryContainer.Resize(fyne.NewSize(250, 60))

		// Получаем прикреплённые дисциплины через gRPC
		conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
		if err != nil {
			log.Printf("Не удалось подключиться к superaccservice: %v", err)
			return
		}
		defer conn.Close()
		client := superaccpb.NewSuperAccServiceClient(conn)
		var attachedDisciplines []string
		resp, err := client.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
		if err == nil {
			for _, g := range resp.Groups {
				if g.Id == group.Id {
					attachedDisciplines = append(attachedDisciplines, g.Disciplines...)
					break
				}
			}
		}

		commentEntry := widget.NewEntry()
		commentEntry.SetText(strings.Join(attachedDisciplines, ", "))
		commentEntry.Disable()
		commentEntryContainer := container.NewMax(commentEntry)
		commentEntryContainer.Resize(fyne.NewSize(250, 60))

		contains := func(s []string, e string) bool {
			for _, a := range s {
				if a == e {
					return true
				}
			}
			return false
		}

		deleteDisciplineButton := widget.NewButton("Удалить дисциплину", func() {
			conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
			if err != nil {
				log.Printf("Не удалось подключиться к superaccservice: %v", err)
				return
			}
			defer conn.Close()
			client := superaccpb.NewSuperAccServiceClient(conn)

			respDisciplines, err := client.ListDisciplines(context.Background(), &superaccpb.ListDisciplinesRequest{})
			if err != nil {
				log.Printf("Не удалось получить список дисциплин: %v", err)
				dialog.ShowInformation("Ошибка", "Не удалось загрузить список дисциплин", w)
				return
			}

			var disciplineOptions []string
			var disciplineIDs []int32
			for _, d := range respDisciplines.Disciplines {
				if contains(attachedDisciplines, d.Name) {
					disciplineOptions = append(disciplineOptions, d.Name)
					disciplineIDs = append(disciplineIDs, d.Id)
				}
			}

			if len(disciplineOptions) == 0 {
				log.Printf("Нет дисциплин для удаления из группы %s", group.Name)
				dialog.ShowInformation("Информация", "Нет прикреплённых дисциплин для удаления", w)
				return
			}

			var checkItems []fyne.CanvasObject
			var checks []*widget.Check
			for _, option := range disciplineOptions {
				check := widget.NewCheck(option, nil)
				checkItems = append(checkItems, check)
				checks = append(checks, check)
			}
			checkGroup := container.NewVBox(checkItems...)

			dialog.ShowForm(
				"Удалить дисциплины",
				"OK",
				"Отмена",
				[]*widget.FormItem{
					widget.NewFormItem("Дисциплины", checkGroup),
				},
				func(confirmed bool) {
					if confirmed {
						var selectedIDs []int32
						for i, check := range checks {
							if check.Checked {
								selectedIDs = append(selectedIDs, disciplineIDs[i])
							}
						}
						if len(selectedIDs) > 0 {
							connFinal, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
							if err != nil {
								log.Printf("Не удалось подключиться к superaccservice: %v", err)
								return
							}
							defer connFinal.Close()
							clientFinal := superaccpb.NewSuperAccServiceClient(connFinal)

							resp, err := clientFinal.DetachDisciplinesFromGroup(context.Background(), &superaccpb.DetachDisciplinesFromGroupRequest{
								GroupId:       group.Id,
								DisciplineIds: selectedIDs,
							})
							if err != nil {
								log.Printf("Не удалось открепить дисциплины: %v", err)
								return
							}
							if !resp.Success {
								log.Printf("Открепление дисциплин не удалось: %s", resp.Message)
							} else {
								log.Printf("Дисциплины успешно откреплены от группы %s", group.Name)
								// Обновляем список дисциплин
								connUpdate, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
								if err == nil {
									defer connUpdate.Close()
									clientUpdate := superaccpb.NewSuperAccServiceClient(connUpdate)
									respUpdate, err := clientUpdate.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
									if err == nil {
										for _, g := range respUpdate.Groups {
											if g.Id == group.Id {
												commentEntry.SetText(strings.Join(g.Disciplines, ", "))
												break
											}
										}
									}
								}
							}
						}
					}
				},
				w,
			)
		})

		attachDisciplineButton := widget.NewButton("Прикрепить дисциплину", func() {
			conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
			if err != nil {
				log.Printf("Не удалось подключиться к superaccservice: %v", err)
				return
			}
			defer conn.Close()
			client := superaccpb.NewSuperAccServiceClient(conn)

			respDisciplines, err := client.ListDisciplines(context.Background(), &superaccpb.ListDisciplinesRequest{})
			if err != nil {
				log.Printf("Не удалось получить список дисциплин: %v", err)
				dialog.ShowInformation("Ошибка", "Не удалось загрузить список дисциплин", w)
				return
			}

			var disciplineOptions []string
			var disciplineIDs []int32
			for _, d := range respDisciplines.Disciplines {
				if !contains(attachedDisciplines, d.Name) {
					disciplineOptions = append(disciplineOptions, d.Name)
					disciplineIDs = append(disciplineIDs, d.Id)
				}
			}

			if len(disciplineOptions) == 0 {
				log.Printf("Нет доступных дисциплин для прикрепления к группе %s", group.Name)
				dialog.ShowInformation("Информация", "Все доступные дисциплины уже прикреплены", w)
				return
			}

			respSeminarists, err := client.GetGroupStaff(context.Background(), &superaccpb.GetGroupStaffRequest{GroupId: group.Id})
			if err != nil {
				log.Printf("Не удалось получить данные о сотрудниках группы: %v", err)
				return
			}
			var seminaristID, assistantID int32
			if respSeminarists.Success {
				seminaristID = respSeminarists.SeminaristId
				assistantID = respSeminarists.AssistantId
			}

			var checkItems []fyne.CanvasObject
			var checks []*widget.Check
			for _, option := range disciplineOptions {
				check := widget.NewCheck(option, nil)
				checkItems = append(checkItems, check)
				checks = append(checks, check)
			}
			checkGroup := container.NewVBox(checkItems...)

			if seminaristID == 0 || assistantID == 0 {
				respUsers, err := client.ListAllUsers(context.Background(), &superaccpb.ListAllUsersRequest{})
				if err != nil {
					log.Printf("Не удалось получить список пользователей: %v", err)
					return
				}
				var userOptions []string
				var userIDs []int32
				for _, u := range respUsers.Users {
					if u.Status == "seminarist" || u.Status == "assistant" {
						userOptions = append(userOptions, fmt.Sprintf("%s (%s)", u.Fio, u.Email))
						userIDs = append(userIDs, u.Id)
					}
				}

				seminaristSelect := widget.NewSelect(append([]string{"None"}, userOptions...), nil)
				assistantSelect := widget.NewSelect(append([]string{"None"}, userOptions...), nil)
				seminaristSelect.SetSelectedIndex(0)
				assistantSelect.SetSelectedIndex(0)

				dialog.ShowForm(
					"Выбор семинариста и ассистента",
					"Далее",
					"Отмена",
					[]*widget.FormItem{
						widget.NewFormItem("Семинарист", seminaristSelect),
						widget.NewFormItem("Ассистент", assistantSelect),
					},
					func(confirmed bool) {
						if confirmed {
							var selectedSeminaristID, selectedAssistantID int32
							if seminaristSelect.SelectedIndex() > 0 {
								selectedSeminaristID = userIDs[seminaristSelect.SelectedIndex()-1]
							}
							if assistantSelect.SelectedIndex() > 0 {
								selectedAssistantID = userIDs[assistantSelect.SelectedIndex()-1]
							}

							// Прикрепляем выбранных семинариста и ассистента к группе
							connInner, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
							if err != nil {
								log.Printf("Не удалось подключиться к superaccservice: %v", err)
								return
							}
							defer connInner.Close()
							clientInner := superaccpb.NewSuperAccServiceClient(connInner)

							_, err = clientInner.ManageGroup(context.Background(), &superaccpb.ManageGroupRequest{
								GroupId: group.Id,
								Action:  "add",
								UserId:  selectedSeminaristID,
								Role:    "seminarist",
							})
							if err != nil {
								log.Printf("Не удалось добавить семинариста: %v", err)
								return
							}
							if selectedAssistantID > 0 {
								_, err = clientInner.ManageGroup(context.Background(), &superaccpb.ManageGroupRequest{
									GroupId: group.Id,
									Action:  "add",
									UserId:  selectedAssistantID,
									Role:    "assistant",
								})
								if err != nil {
									log.Printf("Не удалось добавить ассистента: %v", err)
									return
								}
							}

							var selectedIDs []int32
							dialog.ShowForm(
								"Прикрепить дисциплины",
								"OK",
								"Отмена",
								[]*widget.FormItem{
									widget.NewFormItem("Дисциплины", checkGroup),
								},
								func(confirmed bool) {
									if confirmed {
										for i, check := range checks {
											if check.Checked {
												selectedIDs = append(selectedIDs, disciplineIDs[i])
											}
										}
										if len(selectedIDs) > 0 {
											connFinal, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
											if err != nil {
												log.Printf("Не удалось подключиться к superaccservice: %v", err)
												return
											}
											defer connFinal.Close()
											clientFinal := superaccpb.NewSuperAccServiceClient(connFinal)

											resp, err := clientFinal.ManageDisciplineEntity(context.Background(), &superaccpb.ManageDisciplineEntityRequest{
												Action:        "attach",
												GroupId:       group.Id,
												DisciplineIds: selectedIDs,
												SeminaristId:  selectedSeminaristID,
												AssistantId:   selectedAssistantID,
											})
											if err != nil {
												log.Printf("Не удалось прикрепить дисциплины: %v", err)
												return
											}
											if !resp.Success {
												log.Printf("Прикрепление дисциплин не удалось: %s", resp.Message)
											} else {
												log.Printf("Дисциплины успешно прикреплены к группе %s", group.Name)
												// Обновляем список дисциплин
												connUpdate, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
												if err == nil {
													defer connUpdate.Close()
													clientUpdate := superaccpb.NewSuperAccServiceClient(connUpdate)
													respUpdate, err := clientUpdate.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
													if err == nil {
														for _, g := range respUpdate.Groups {
															if g.Id == group.Id {
																commentEntry.SetText(strings.Join(g.Disciplines, ", "))
																break
															}
														}
													}
												}
											}
										}
									}
								},
								w,
							)
						}
					},
					w,
				)
			} else {
				var selectedIDs []int32
				dialog.ShowForm(
					"Прикрепить дисциплины",
					"OK",
					"Отмена",
					[]*widget.FormItem{
						widget.NewFormItem("Дисциплины", checkGroup),
					},
					func(confirmed bool) {
						if confirmed {
							for i, check := range checks {
								if check.Checked {
									selectedIDs = append(selectedIDs, disciplineIDs[i])
								}
							}
							if len(selectedIDs) > 0 {
								connFinal, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
								if err != nil {
									log.Printf("Не удалось подключиться к superaccservice: %v", err)
									return
								}
								defer connFinal.Close()
								clientFinal := superaccpb.NewSuperAccServiceClient(connFinal)

								resp, err := clientFinal.ManageDisciplineEntity(context.Background(), &superaccpb.ManageDisciplineEntityRequest{
									Action:        "attach",
									GroupId:       group.Id,
									DisciplineIds: selectedIDs,
									SeminaristId:  seminaristID,
									AssistantId:   assistantID,
								})
								if err != nil {
									log.Printf("Не удалось прикрепить дисциплины: %v", err)
									return
								}
								if !resp.Success {
									log.Printf("Прикрепление дисциплин не удалось: %s", resp.Message)
								} else {
									log.Printf("Дисциплины успешно прикреплены к группе %s", group.Name)
									// Обновляем список дисциплин
									connUpdate, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
									if err == nil {
										defer connUpdate.Close()
										clientUpdate := superaccpb.NewSuperAccServiceClient(connUpdate)
										respUpdate, err := clientUpdate.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
										if err == nil {
											for _, g := range respUpdate.Groups {
												if g.Id == group.Id {
													commentEntry.SetText(strings.Join(g.Disciplines, ", "))
													break
												}
											}
										}
									}
								}
							}
						}
					},
					w,
				)
			}
		})

		nextButton := widget.NewButton("Подробнее", func() {
			log.Printf("Кнопка 'Подробнее' нажата для группы ID: %d", group.Id)
			GroupName = group.Name
			state.currentPage = "superacc-users-of-group"
			w.SetContent(createContent(state))
		})

		deleteButton := widget.NewButton("Удалить", func() {
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Удалить группу '%s'?", group.Name),
				func(confirmed bool) {
					if confirmed {
						conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
						if err != nil {
							log.Printf("Не удалось подключиться к superaccservice: %v", err)
							return
						}
						defer conn.Close()

						client := superaccpb.NewSuperAccServiceClient(conn)
						resp, err := client.ManageGroupEntity(context.Background(), &superaccpb.ManageGroupEntityRequest{
							GroupId: group.Id,
							Action:  "delete",
						})
						if err != nil {
							log.Printf("Не удалось удалить группу: %v", err)
							return
						}
						if !resp.Success {
							log.Printf("Удаление группы не удалось: %s", resp.Message)
						} else {
							log.Printf("Группа %s с ID %d успешно удалена", group.Name, group.Id)
							for i, g := range groups {
								if g.NameEntry.Text == group.Name {
									groupInfoListContainer.Remove(g.Container)
									groups = append(groups[:i], groups[i+1:]...)
									groupInfoListContainer.Refresh()
									break
								}
							}
						}
					}
				},
				w,
			)
		})

		groupRow := container.New(layout.NewGridLayoutWithColumns(7),
			container.NewPadded(container.NewPadded(nameEntryContainer)),
			container.NewPadded(container.NewPadded(descriptionEntryContainer)),
			container.NewPadded(container.NewPadded(commentEntryContainer)),
			container.NewPadded(container.NewPadded(deleteDisciplineButton)),
			container.NewPadded(container.NewPadded(attachDisciplineButton)),
			container.NewPadded(container.NewPadded(nextButton)),
			container.NewPadded(container.NewPadded(deleteButton)),
		)

		groupInfoListContainer.Add(groupRow)
		groupInfoListContainer.Refresh()

		groups = append(groups, &GroupEntry{
			NameEntry:        nameEntry,
			DescriptionEntry: descriptionEntry,
			CommentEntry:     commentEntry,
			Container:        groupRow,
		})
	}

	conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to superaccservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу"))
	}
	defer conn.Close()

	client := superaccpb.NewSuperAccServiceClient(conn)
	resp, err := client.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
	if err != nil {
		log.Printf("Failed to list groups: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки групп"))
	}
	if !resp.Success {
		log.Printf("ListGroups failed: %s", resp.Message)
		return container.NewVBox(widget.NewLabel(fmt.Sprintf("Ошибка: %s", resp.Message)))
	}

	for _, group := range resp.Groups {
		addCriterionEntry(group)
	}

	listLabel := widget.NewLabelWithStyle("Список групп", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	scrollableCriteria := container.NewVScroll(groupInfoListContainer)
	scrollableCriteria.SetMinSize(fyne.NewSize(0, 400))

	addButton := widget.NewButton("Добавить", func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Название группы")
		descriptionEntry := widget.NewEntry()
		descriptionEntry.SetPlaceHolder("Описание")

		dialog.ShowForm(
			"Добавить группу",
			"OK",
			"Отмена",
			[]*widget.FormItem{
				widget.NewFormItem("Название", nameEntry),
				widget.NewFormItem("Описание", descriptionEntry),
			},
			func(confirmed bool) {
				if confirmed && nameEntry.Text != "" {
					conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
					if err != nil {
						log.Printf("Failed to connect to superaccservice: %v", err)
						return
					}
					defer conn.Close()

					client := superaccpb.NewSuperAccServiceClient(conn)
					resp, err := client.ManageGroupEntity(context.Background(), &superaccpb.ManageGroupEntityRequest{
						Name:        nameEntry.Text,
						Description: descriptionEntry.Text,
						Action:      "create",
					})
					if err != nil {
						log.Printf("Failed to create group: %v", err)
						return
					}
					if !resp.Success {
						log.Printf("Create group failed: %s", resp.Message)
						return
					}

					newResp, err := client.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
					if err != nil {
						log.Printf("Failed to refresh groups: %v", err)
						return
					}
					groupInfoListContainer.RemoveAll()
					groups = nil
					for _, group := range newResp.Groups {
						addCriterionEntry(group)
					}
					groupInfoListContainer.Refresh()
					log.Printf("Группа с ID %d успешно добавлена", resp.GroupId)
				}
			},
			w,
		)
	})

	nextButton := widget.NewButton("Список учеников", func() {
		fmt.Println("Кнопка 'Далее' нажата. Собираем данные групп:")
		state.currentPage = "superacc-all-users"
		w.SetContent(createContent(state))
		return
	})

	createDisciplineButton := widget.NewButton("Создать дисциплину", func() {
		conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to superaccservice: %v", err)
			dialog.ShowInformation("Ошибка", "Не удалось подключиться к серверу", w)
			return
		}
		defer conn.Close()
		log.Printf("Connection established, creating client")
		client := superaccpb.NewSuperAccServiceClient(conn)

		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Название дисциплины")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		respUsers, err := client.ListAllUsers(ctx, &superaccpb.ListAllUsersRequest{})
		if err != nil {
			log.Printf("Failed to list users: %v", err)
			dialog.ShowInformation("Ошибка", "Не удалось загрузить список пользователей", w)
			return
		}
		var lecturerOptions []string
		var lecturerIDs []int32
		for _, u := range respUsers.Users {
			if u.Status == "lecturer" {
				lecturerOptions = append(lecturerOptions, fmt.Sprintf("%s (%s)", u.Fio, u.Email))
				lecturerIDs = append(lecturerIDs, u.Id)
			}
		}
		lecturerSelect := widget.NewSelect(append([]string{"None"}, lecturerOptions...), nil)
		lecturerSelect.SetSelectedIndex(0)

		dialog.ShowForm(
			"Создать дисциплину",
			"OK",
			"Отмена",
			[]*widget.FormItem{
				widget.NewFormItem("Название", nameEntry),
				widget.NewFormItem("Лектор", lecturerSelect),
			},
			func(confirmed bool) {
				if confirmed && nameEntry.Text != "" {
					var lectorID int32
					if lecturerSelect.SelectedIndex() > 0 {
						lectorID = lecturerIDs[lecturerSelect.SelectedIndex()-1]
					} else {
						dialog.ShowInformation("Ошибка", "Не выбран лектор", w)
						return
					}

					conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(15*time.Second))
					if err != nil {
						log.Printf("Failed to connect to superaccservice: %v", err)
						dialog.ShowInformation("Ошибка", "Не удалось подключиться к серверу", w)
						return
					}
					defer conn.Close()
					log.Printf("Connection established, creating client")
					client := superaccpb.NewSuperAccServiceClient(conn)

					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					log.Printf("Sending CreateDiscipline request for %s with lectorID %d", nameEntry.Text, lectorID)
					resp, err := client.CreateDiscipline(ctx, &superaccpb.ManageDisciplineEntityRequest{
						Action:   "create",
						Name:     nameEntry.Text,
						LectorId: lectorID,
					})
					if err != nil {
						log.Printf("Не удалось создать дисциплину: %v", err)
						dialog.ShowInformation("Ошибка", "Не удалось создать дисциплину", w)
						return
					}
					if !resp.Success {
						log.Printf("Ошибка создания дисциплины: %s", resp.Message)
						dialog.ShowInformation("Ошибка", resp.Message, w)
						return
					}
					log.Printf("Дисциплина успешно создана с ID %d", resp.DisciplineId)
					dialog.ShowInformation("Успех", fmt.Sprintf("Дисциплина '%s' создана с ID %d", nameEntry.Text, resp.DisciplineId), w)
				}
			},
			w,
		)
	})

	deleteDisciplineButton := widget.NewButton("Удалить дисциплину", func() {
		client := superaccpb.NewSuperAccServiceClient(nil) // Инициализация без соединения
		var conn *grpc.ClientConn

		// Получаем список дисциплин внутри диалога
		conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to superaccservice: %v", err)
			dialog.ShowInformation("Ошибка", "Не удалось подключиться к серверу", w)
			return
		}
		defer conn.Close()
		log.Printf("Connection established, creating client")
		client = superaccpb.NewSuperAccServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		respDisciplines, err := client.ListDisciplines(ctx, &superaccpb.ListDisciplinesRequest{})
		if err != nil {
			log.Printf("Не удалось получить список дисциплин: %v", err)
			dialog.ShowInformation("Ошибка", "Не удалось загрузить список дисциплин", w)
			return
		}

		var disciplineOptions []string
		var disciplineIDs []int32
		for _, d := range respDisciplines.Disciplines {
			disciplineOptions = append(disciplineOptions, d.Name)
			disciplineIDs = append(disciplineIDs, d.Id)
		}

		if len(disciplineOptions) == 0 {
			log.Printf("Нет дисциплин для удаления")
			dialog.ShowInformation("Информация", "Нет дисциплин для удаления", w)
			return
		}

		var checkItems []fyne.CanvasObject
		var checks []*widget.Check
		for _, option := range disciplineOptions {
			check := widget.NewCheck(option, nil)
			checkItems = append(checkItems, check)
			checks = append(checks, check)
		}
		checkGroup := container.NewVBox(checkItems...)

		dialog.ShowForm(
			"Удалить дисциплины",
			"OK",
			"Отмена",
			[]*widget.FormItem{widget.NewFormItem("Дисциплины", checkGroup)},
			func(confirmed bool) {
				if confirmed {
					var selectedIDs []int32
					for i, check := range checks {
						if check.Checked {
							selectedIDs = append(selectedIDs, disciplineIDs[i])
						}
					}
					if len(selectedIDs) > 0 {
						// Создаем новое соединение для удаления
						conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
						if err != nil {
							log.Printf("Failed to connect to superaccservice: %v", err)
							dialog.ShowInformation("Ошибка", "Не удалось подключиться к серверу", w)
							return
						}
						defer conn.Close()
						log.Printf("Connection established, creating client")
						client := superaccpb.NewSuperAccServiceClient(conn)

						ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
						defer cancel()
						log.Printf("Sending DeleteDiscipline request for IDs %v", selectedIDs)
						resp, err := client.DeleteDiscipline(ctx, &superaccpb.DeleteDisciplineRequest{
							DisciplineIds: selectedIDs,
						})
						if err != nil {
							log.Printf("Не удалось удалить дисциплины: %v", err)
							dialog.ShowInformation("Ошибка", "Не удалось удалить дисциплины", w)
							return
						}
						if !resp.Success {
							log.Printf("Удаление дисциплин не удалось: %s", resp.Message)
							dialog.ShowInformation("Ошибка", resp.Message, w)
							return
						}
						log.Printf("Дисциплины удалены успешно: %v", selectedIDs)
						dialog.ShowInformation("Успех", fmt.Sprintf("Дисциплины с ID %v удалены успешно", selectedIDs), w)
					}
				}
			},
			w,
		)
	})

	bottomButtons := container.New(layout.NewHBoxLayout(),
		addButton,
		layout.NewSpacer(),
		deleteDisciplineButton,
		createDisciplineButton,
		nextButton,
	)

	bottomButtonsWithPadding := container.NewPadded(bottomButtons)

	contentBackground := canvas.NewRectangle(color.White)

	criteriaPanel := container.NewVBox(
		container.NewPadded(columnHeaders),
		listLabel,
		scrollableCriteria,
	)

	centralContent := container.NewStack(
		contentBackground,
		container.NewPadded(criteriaPanel),
	)

	return container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			container.NewVBox(headerContent, backButtonContainer),
			bottomButtonsWithPadding,
			nil,
			nil,
			centralContent,
		),
	)
}

//****************************************
//* Page with all users of current group *
//****************************************

type UserEntry struct {
	FIOEmail string
	Status   string
}

func СreateGroupUsersPage(state *AppState, groupName string) fyne.CanvasObject {
	w := state.window
	logoText := canvas.NewText("ВШЭ", color.White)
	logoText.TextStyle.Bold = true
	logoText.TextSize = 24
	logoText.Alignment = fyne.TextAlignCenter
	leftHeaderObject := container.NewStack(logoText)

	headerTitle := canvas.NewText(fmt.Sprintf("Список пользователей: %s", groupName), color.White)
	headerTitle.TextStyle.Bold = true
	headerTitle.TextSize = 20
	headerTitle.Alignment = fyne.TextAlignCenter

	headerContent := container.New(layout.NewBorderLayout(nil, nil, leftHeaderObject, nil),
		leftHeaderObject,
		headerTitle,
	)

	backButton := widget.NewButton("Назад", func() {
		fmt.Println("Кнопка 'Назад' нажата. Возврат на предыдущую страницу.")
		state.currentPage = "superacc-groups"
		w.SetContent(createContent(state))
	})
	backButtonRow := container.NewHBox(backButton, layout.NewSpacer())

	usersListContainer := container.NewVBox()

	headerFIOEmail := widget.NewLabelWithStyle("ФИО, почта", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerStatus := widget.NewLabelWithStyle("Статус", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	headerCellBackground := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	verticalHeaderDivider := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
	verticalHeaderDivider.SetMinSize(fyne.NewSize(1, 0))

	columnHeaders := container.New(layout.NewHBoxLayout(),
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerFIOEmail))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerStatus))),
	)
	columnHeadersContainer := container.New(layout.NewHBoxLayout(), columnHeaders, layout.NewSpacer())

	var usersData []UserEntry

	var updateUsersListUI func()

	extractEmail := func(fioEmail string) string {
		parts := strings.Split(fioEmail, ", ")
		if len(parts) > 1 {
			return parts[1]
		}
		return ""
	}

	createUserRow := func(user UserEntry, idx int) *fyne.Container {
		fioEmailLabel := widget.NewLabel(user.FIOEmail)
		fioEmailLabel.Wrapping = fyne.TextWrapWord

		statusOptions := []string{"student", "assistant", "seminarist", "lecturer", "superaccount"}
		statusSelect := widget.NewSelect(statusOptions, func(selected string) {
			usersData[idx].Status = selected
			fmt.Printf("Статус пользователя %s изменен на: %s\n", usersData[idx].FIOEmail, selected)

			conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to superaccservice: %v", err)
				return
			}
			defer conn.Close()

			client := superaccpb.NewSuperAccServiceClient(conn)
			resp, err := client.ListAllUsers(context.Background(), &superaccpb.ListAllUsersRequest{})
			if err != nil {
				log.Printf("Failed to list users: %v", err)
				return
			}
			var userID int32
			email := extractEmail(user.FIOEmail)
			for _, u := range resp.Users {
				if u.Email == email {
					userID = u.Id
					break
				}
			}
			if userID == 0 {
				log.Printf("User with email %s not found", email)
				return
			}

			updateResp, err := client.UpdateUserRole(context.Background(), &superaccpb.UpdateRoleRequest{
				UserId: userID,
				Role:   selected,
			})
			if err != nil {
				log.Printf("Failed to update role: %v", err)
				return
			}
			if !updateResp.Success {
				log.Printf("Update role failed: %s", updateResp.Message)
			} else {
				log.Printf("Role updated successfully for user %s", user.FIOEmail)
			}
		})
		statusSelect.SetSelected(user.Status)

		deleteButton := widget.NewButton("Удалить", func() {
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Удалить пользователя '%s' из группы?", user.FIOEmail),
				func(confirmed bool) {
					if confirmed {
						conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
						if err != nil {
							log.Printf("Failed to connect to superaccservice: %v", err)
							return
						}
						defer conn.Close()

						client := superaccpb.NewSuperAccServiceClient(conn)
						email := extractEmail(user.FIOEmail)

						respUsers, err := client.ListAllUsers(context.Background(), &superaccpb.ListAllUsersRequest{})
						if err != nil {
							log.Printf("Failed to list users: %v", err)
							return
						}
						var userID int32
						for _, u := range respUsers.Users {
							if u.Email == email {
								userID = u.Id
								break
							}
						}
						if userID == 0 {
							log.Printf("User with email %s not found", email)
							return
						}

						respGroups, err := client.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
						if err != nil {
							log.Printf("Failed to list groups: %v", err)
							return
						}
						var groupID int32
						for _, g := range respGroups.Groups {
							if g.Name == groupName {
								groupID = g.Id
								break
							}
						}
						if groupID == 0 {
							log.Printf("Group %s not found", groupName)
							return
						}

						resp, err := client.ManageGroup(context.Background(), &superaccpb.ManageGroupRequest{
							GroupId: groupID,
							Action:  "remove",
							UserId:  userID,
							Role:    user.Status,
						})
						if err != nil {
							log.Printf("Failed to remove user from group: %v", err)
							return
						}
						if !resp.Success {
							log.Printf("Remove user from group failed: %s", resp.Message)
						} else {
							log.Printf("User %s removed from group %s", user.FIOEmail, groupName)
							respUsersByGroup, err := client.ListUsersByGroup(context.Background(), &superaccpb.ListUsersByGroupRequest{GroupName: groupName})
							if err == nil {
								usersData = make([]UserEntry, len(respUsersByGroup.Users))
								for i, user := range respUsersByGroup.Users {
									usersData[i] = UserEntry{FIOEmail: user.Fio + ", " + user.Email, Status: user.Status}
								}
								updateUsersListUI()
							} else {
								log.Printf("Failed to refresh users: %v", err)
							}
						}
					}
				},
				w,
			)
		})

		cellFIOEmail := container.NewPadded(container.NewMax(fioEmailLabel))
		cellStatus := container.NewPadded(container.NewMax(statusSelect))
		cellDelete := container.NewPadded(container.NewMax(deleteButton))

		verticalCellDivider1 := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		verticalCellDivider1.SetMinSize(fyne.NewSize(1, 0))
		verticalCellDivider2 := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		verticalCellDivider2.SetMinSize(fyne.NewSize(1, 0))

		rowContainer := container.New(layout.NewVBoxLayout(),
			cellFIOEmail,
			verticalCellDivider1,
			cellStatus,
			verticalCellDivider2,
			cellDelete,
		)
		return rowContainer
	}

	updateUsersListUI = func() {
		usersListContainer.RemoveAll()

		if len(usersData) == 0 {
			usersListContainer.Add(container.NewCenter(widget.NewLabel("Нет пользователей для отображения")))
			usersListContainer.Refresh()
			return
		}

		for i, user := range usersData {
			usersListContainer.Add(createUserRow(user, i))
			usersListContainer.Add(canvas.NewRectangle(color.NRGBA{R: 220, G: 220, B: 220, A: 255}))
		}
		usersListContainer.Refresh()
	}

	conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to superaccservice: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка подключения к серверу"))
	}
	defer conn.Close()

	client := superaccpb.NewSuperAccServiceClient(conn)
	resp, err := client.ListUsersByGroup(context.Background(), &superaccpb.ListUsersByGroupRequest{GroupName: groupName})
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return container.NewVBox(widget.NewLabel("Ошибка загрузки пользователей"))
	}
	if !resp.Success {
		log.Printf("ListUsersByGroup failed: %s", resp.Message)
		return container.NewVBox(widget.NewLabel(fmt.Sprintf("Ошибка: %s", resp.Message)))
	}

	usersData = make([]UserEntry, len(resp.Users))
	for i, user := range resp.Users {
		usersData[i] = UserEntry{FIOEmail: user.Fio + ", " + user.Email, Status: user.Status}
	}

	updateUsersListUI()

	scrollableUsersList := container.NewVScroll(usersListContainer)
	scrollableUsersList.SetMinSize(fyne.NewSize(215, 450))

	addButton := widget.NewButton("Добавить", func() {
		searchEntry := widget.NewEntry()
		searchEntry.SetPlaceHolder("Введите имя для поиска...")
		var filteredUsers []UserEntry

		updateFilteredUsers := func() {
			query := strings.ToLower(searchEntry.Text)
			filteredUsers = nil

			conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to superaccservice: %v", err)
				return
			}
			defer conn.Close()

			client := superaccpb.NewSuperAccServiceClient(conn)
			allUsersResp, err := client.ListAllUsers(context.Background(), &superaccpb.ListAllUsersRequest{})
			if err != nil {
				log.Printf("Failed to list all users: %v", err)
				return
			}

			for _, user := range allUsersResp.Users {
				if strings.Contains(strings.ToLower(user.Fio), query) || strings.Contains(strings.ToLower(user.Email), query) {
					if user.Group == groupName || user.Group == "" {
						filteredUsers = append(filteredUsers, UserEntry{FIOEmail: user.Fio + ", " + user.Email, Status: user.Status})
					}
				}
			}
		}

		userList := widget.NewList(
			func() int { return len(filteredUsers) },
			func() fyne.CanvasObject {
				return widget.NewLabel("ФИО, почта")
			},
			func(id widget.ListItemID, item fyne.CanvasObject) {
				item.(*widget.Label).SetText(filteredUsers[id].FIOEmail)
			},
		)

		userList.OnSelected = func(id widget.ListItemID) {
			selectedUser := filteredUsers[id]
			conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to superaccservice: %v", err)
				return
			}
			defer conn.Close()

			client := superaccpb.NewSuperAccServiceClient(conn)
			email := extractEmail(selectedUser.FIOEmail)

			respUsers, err := client.ListAllUsers(context.Background(), &superaccpb.ListAllUsersRequest{})
			if err != nil {
				log.Printf("Failed to list users: %v", err)
				return
			}
			var userID int32
			for _, user := range respUsers.Users {
				if user.Email == email {
					userID = user.Id
					break
				}
			}
			if userID == 0 {
				log.Printf("User with email %s not found", email)
				return
			}

			var groupID int32
			respGroups, err := client.ListGroups(context.Background(), &superaccpb.ListGroupsRequest{})
			if err != nil {
				log.Printf("Failed to list groups: %v", err)
				return
			}
			for _, g := range respGroups.Groups {
				if g.Name == groupName {
					groupID = g.Id
					break
				}
			}
			if groupID == 0 {
				log.Printf("Group %s not found", groupName)
				return
			}

			resp, err := client.ManageGroup(context.Background(), &superaccpb.ManageGroupRequest{
				GroupId: groupID,
				Action:  "add",
				UserId:  userID,
				Role:    selectedUser.Status,
			})
			if err != nil {
				log.Printf("Failed to add user to group: %v", err)
				dialog.ShowInformation("Ошибка", fmt.Sprintf("Не удалось прикрепить пользователя: %v", err), w)
				return
			}
			if !resp.Success {
				log.Printf("Manage group failed: %s", resp.Message)
				dialog.ShowInformation("Ошибка", fmt.Sprintf("Не удалось прикрепить пользователя: %s", resp.Message), w)
				return
			}

			respUsersByGroup, err := client.ListUsersByGroup(context.Background(), &superaccpb.ListUsersByGroupRequest{GroupName: groupName})
			if err != nil {
				log.Printf("Failed to refresh users: %v", err)
				return
			}
			usersData = make([]UserEntry, len(respUsersByGroup.Users))
			for i, user := range respUsersByGroup.Users {
				usersData[i] = UserEntry{FIOEmail: user.Fio + ", " + user.Email, Status: user.Status}
			}
			updateUsersListUI()
			log.Printf("Пользователь %s прикреплён к группе %s", selectedUser.FIOEmail, groupName)
		}

		searchEntry.OnChanged = func(s string) {
			updateFilteredUsers()
			userList.Refresh()
		}
		updateFilteredUsers()

		dialogContent := container.NewVBox(
			searchEntry,
			container.NewVScroll(userList),
		)

		dialog.ShowCustom("Поиск пользователя", "Закрыть", dialogContent, w)
	})

	contentBackground := canvas.NewRectangle(color.White)
	usersPanel := container.NewVBox(
		backButtonRow,
		columnHeadersContainer,
		scrollableUsersList,
		addButton,
	)
	centralContent := container.NewStack(
		contentBackground,
		container.NewPadded(usersPanel),
	)

	return container.NewStack(
		canvas.NewRectangle(color.NRGBA{R: 20, G: 40, B: 80, A: 255}),
		container.NewBorder(
			headerContent,
			nil,
			nil,
			nil,
			centralContent,
		),
	)
}

//***********************************
//* Page with all registrated users *
//***********************************

type User struct {
	ID     int
	FIO    string
	Email  string
	Group  string
	Status string
}

func СreateUsersListPage(state *AppState) fyne.CanvasObject {
	w := state.window

	var currentDisplayedUsers []*User
	var updateUsersTableUI func(searchText string)

	headerTextColor := color.White
	darkBlue := color.NRGBA{R: 20, G: 40, B: 80, A: 255}
	lightGrayDivider := color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	mediumGrayDivider := color.NRGBA{R: 180, G: 180, B: 180, A: 255}

	logo := canvas.NewText("ВШЭ", headerTextColor)
	logo.TextStyle.Bold = true
	logo.TextSize = 28
	logoContainer := container.NewMax(logo)

	backButton := widget.NewButton("Назад", func() {
		log.Printf("Кнопка 'Назад' нажата. Возврат на предыдущую страницу.")
		state.currentPage = "superacc-groups"
		w.SetContent(createContent(state))
	})
	backButtonContainer := container.NewVBox(backButton)

	headerTitleText := canvas.NewText("Список пользователей", headerTextColor)
	headerTitleText.TextStyle.Bold = true
	headerTitleText.TextSize = 24
	headerTitleText.Alignment = fyne.TextAlignCenter

	header := container.New(layout.NewBorderLayout(backButtonContainer, nil, logoContainer, nil),
		backButtonContainer,
		container.NewPadded(logoContainer),
		container.NewCenter(headerTitleText),
	)
	headerBackground := canvas.NewRectangle(darkBlue)
	headerWithBackground := container.NewStack(headerBackground, header)

	sidePanelBackground := canvas.NewRectangle(darkBlue)
	sidePanel := container.NewVBox()
	sidePanelWithBackground := container.NewStack(sidePanelBackground, sidePanel)

	tableRowsContainer := container.NewVBox()
	scrollableTable := container.NewVScroll(tableRowsContainer)
	scrollableTable.SetMinSize(fyne.NewSize(0, 450))

	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "поиск                                     "
	searchEntry.OnChanged = func(s string) {
		updateUsersTableUI(s)
	}
	searchBox := container.New(layout.NewVBoxLayout(),
		widget.NewLabel("поиск"),
		searchEntry,
		layout.NewSpacer(),
	)

	headerFIOEmail := widget.NewLabelWithStyle("ФИО почта", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerGroup := widget.NewLabelWithStyle("группа", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	headerStatus := widget.NewLabelWithStyle("Статус", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	headerCellBackground := canvas.NewRectangle(color.White)

	verticalHeaderDivider := canvas.NewRectangle(mediumGrayDivider)
	verticalHeaderDivider.SetMinSize(fyne.NewSize(1, 0))

	columnHeaders := container.New(layout.NewHBoxLayout(),
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerFIOEmail))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerGroup))),
		verticalHeaderDivider,
		container.NewMax(headerCellBackground, container.NewPadded(container.NewCenter(headerStatus))),
	)
	columnHeadersContainer := container.New(layout.NewHBoxLayout(), columnHeaders, layout.NewSpacer())

	createUserTableRow := func(user *User, idx int) *fyne.Container {
		fioEmailCombinedLabel := widget.NewLabel(fmt.Sprintf("%s, %s", user.FIO, user.Email))
		fioEmailCombinedLabel.Wrapping = fyne.TextWrapWord

		groupLabel := widget.NewLabel(user.Group)
		groupLabel.Wrapping = fyne.TextWrapWord

		statusOptions := []string{"student", "assistant", "seminarist", "lecturer", "superaccount"}
		statusSelect := widget.NewSelect(statusOptions, func(selectedStatus string) {
			user.Status = selectedStatus
			fmt.Printf("Статус пользователя %s (%s) изменен на: %s\n", user.FIO, user.Email, selectedStatus)

			conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to superaccservice: %v", err)
				return
			}
			defer conn.Close()

			client := superaccpb.NewSuperAccServiceClient(conn)
			resp, err := client.UpdateUserRole(context.Background(), &superaccpb.UpdateRoleRequest{
				UserId: int32(user.ID),
				Role:   selectedStatus,
			})
			if err != nil {
				log.Printf("Failed to update role: %v", err)
				return
			}
			if !resp.Success {
				log.Printf("Update role failed: %s", resp.Message)
			} else {
				log.Printf("Role updated successfully for user %s", user.FIO)
			}
		})
		statusSelect.SetSelected(user.Status)

		deleteButton := widget.NewButton("Удалить", func() {
			dialog.ShowConfirm(
				"Подтверждение удаления",
				fmt.Sprintf("Удалить пользователя '%s (%s)'?", user.FIO, user.Email),
				func(confirmed bool) {
					if confirmed {
						conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
						if err != nil {
							log.Printf("Failed to connect to superaccservice: %v", err)
							return
						}
						defer conn.Close()

						client := superaccpb.NewSuperAccServiceClient(conn)
						resp, err := client.RemoveUser(context.Background(), &superaccpb.RemoveUserRequest{
							Email: user.Email,
						})
						if err != nil {
							log.Printf("Failed to remove user: %v", err)
							return
						}
						if !resp.Success {
							log.Printf("Remove user failed: %s", resp.Message)
						} else {
							log.Printf("User %s (%s) successfully removed", user.FIO, user.Email)
							updateUsersTableUI(searchEntry.Text)
						}
					}
				},
				w,
			)
		})

		cellFIOEmail := container.NewPadded(container.NewMax(fioEmailCombinedLabel))
		cellGroup := container.NewPadded(container.NewMax(groupLabel))
		cellStatus := container.NewPadded(container.NewMax(statusSelect))
		cellDelete := container.NewPadded(container.NewMax(deleteButton))

		verticalCellDivider := canvas.NewRectangle(mediumGrayDivider)
		verticalCellDivider.SetMinSize(fyne.NewSize(1, 0))

		rowContainer := container.New(layout.NewVBoxLayout(),
			cellFIOEmail,
			verticalCellDivider,
			cellGroup,
			verticalCellDivider,
			cellStatus,
			cellDelete,
		)
		return rowContainer
	}

	updateUsersTableUI = func(searchText string) {
		tableRowsContainer.RemoveAll()

		conn, err := grpc.Dial("89.169.39.161:50052", grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to superaccservice: %v", err)
			return
		}
		defer conn.Close()

		client := superaccpb.NewSuperAccServiceClient(conn)
		resp, err := client.ListAllUsers(context.Background(), &superaccpb.ListAllUsersRequest{})
		if err != nil {
			log.Printf("Failed to list all users: %v", err)
			return
		}

		currentDisplayedUsers = []*User{}
		if searchText == "" {
			currentDisplayedUsers = make([]*User, len(resp.Users))
			for i, user := range resp.Users {
				currentDisplayedUsers[i] = &User{
					ID:     int(user.Id),
					FIO:    user.Fio,
					Email:  user.Email,
					Group:  user.Group,
					Status: user.Status,
				}
			}
		} else {
			lowerSearchText := strings.ToLower(searchText)
			for _, user := range resp.Users {
				if strings.Contains(strings.ToLower(user.Fio), lowerSearchText) ||
					strings.Contains(strings.ToLower(user.Email), lowerSearchText) {
					currentDisplayedUsers = append(currentDisplayedUsers, &User{
						ID:     int(user.Id),
						FIO:    user.Fio,
						Email:  user.Email,
						Group:  user.Group,
						Status: user.Status,
					})
				}
			}
		}

		if len(currentDisplayedUsers) == 0 {
			tableRowsContainer.Add(container.NewCenter(widget.NewLabel("Нет пользователей для отображения по заданным критериям.")))
		} else {
			for i, user := range currentDisplayedUsers {
				tableRowsContainer.Add(createUserTableRow(user, i))
				tableRowsContainer.Add(canvas.NewRectangle(lightGrayDivider))
			}
		}
		tableRowsContainer.Refresh()
		scrollableTable.ScrollToTop()
	}

	updateUsersTableUI("")

	centralContentPanel := container.NewVBox(
		container.NewPadded(searchBox),
		columnHeadersContainer,
		scrollableTable,
	)

	contentBackground := canvas.NewRectangle(color.White)
	centralContentWithBackground := container.NewStack(contentBackground, container.NewPadded(centralContentPanel))

	return container.NewBorder(
		headerWithBackground,
		nil,
		sidePanelWithBackground,
		nil,
		centralContentWithBackground,
	)
}
