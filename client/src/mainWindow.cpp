#include "mainWindow.h"

#include "apiClient.h"
#include "deviceDialog.h"
#include "vendorsDialog.h"
#include "modelsDialog.h"
#include "locationsDialog.h"
#include "usersDialog.h"
#include "themeManager.h"
#include "messageBoxUtils.h"

#include <QLabel>
#include <QLineEdit>
#include <QMessageBox>
#include <QTableWidget>
#include <QHeaderView>
#include <QToolBar>
#include <QVBoxLayout>
#include <QWidget>

MainWindow::MainWindow(ApiClient* apiClient, const QString& username, const QString& role, QWidget* parent)
    : QMainWindow(parent)
    , apiClient_(apiClient)
    , devices_()
    , searchEdit_(nullptr)
    , table_(nullptr)
    , addAction_(nullptr)
    , editAction_(nullptr)
    , vendorsAction_(nullptr)
    , modelsAction_(nullptr)
    , locationsAction_(nullptr)
    , usersAction_(nullptr)
    , refreshAction_(nullptr)
    , deleteAction_(nullptr)
    , logoutAction_(nullptr) {
    setWindowTitle("TelecomBase");
    buildUi(username, role);
    loadDevices();
}

void MainWindow::setSession(const QString& username, const QString& role) {
    // Clear existing UI
    const auto toolbars = findChildren<QToolBar*>();
    for (auto* tb : toolbars) {
        removeToolBar(tb);
        delete tb;
    }

    if (auto* cw = centralWidget()) {
        setCentralWidget(nullptr);
        delete cw;
    }

    // Reset pointers
    devices_.clear();
    searchEdit_ = nullptr;
    table_ = nullptr;
    addAction_ = nullptr;
    editAction_ = nullptr;
    vendorsAction_ = nullptr;
    modelsAction_ = nullptr;
    locationsAction_ = nullptr;
    usersAction_ = nullptr;
    refreshAction_ = nullptr;
    deleteAction_ = nullptr;
    logoutAction_ = nullptr;
    themeAction_ = nullptr;

    buildUi(username, role);
    loadDevices();
}

void MainWindow::buildUi(const QString& username, const QString& role) {
    QToolBar* toolbar = addToolBar("Действия");
    toolbar->setMovable(false);
    addAction_ = toolbar->addAction("Добавить");
    editAction_ = toolbar->addAction("Редактировать");
    deleteAction_ = toolbar->addAction("Удалить");
    if (deleteAction_) {
        deleteAction_->setEnabled(role == "admin");
        if (role != "admin") {
            deleteAction_->setToolTip("Удаление доступно только admin");
        }
    }
    toolbar->addSeparator();
    refreshAction_ = toolbar->addAction("Обновить");

    toolbar->addSeparator();
    logoutAction_ = toolbar->addAction("Выйти");

    toolbar->addSeparator();
    if (role == "admin") {
        vendorsAction_ = toolbar->addAction("Производители");
        modelsAction_ = toolbar->addAction("Модели");
        locationsAction_ = toolbar->addAction("Локации");
        usersAction_ = toolbar->addAction("Пользователи");
        toolbar->addSeparator();
    }

    themeAction_ = toolbar->addAction("Тёмная тема");
    themeAction_->setCheckable(true);
    themeAction_->setChecked(ThemeManager::theme() == ThemeManager::Theme::Dark);

    QWidget* central = new QWidget(this);
    QVBoxLayout* layout = new QVBoxLayout(central);
    layout->setContentsMargins(16, 16, 16, 16);
    layout->setSpacing(12);

    QLabel* userInfo = new QLabel(QString("Пользователь: %1 (%2)").arg(username, role), central);
    layout->addWidget(userInfo);

    searchEdit_ = new QLineEdit(central);
    searchEdit_->setPlaceholderText("Поиск (серийный/инвентарный, модель, производитель, статус)");
    layout->addWidget(searchEdit_);

    table_ = new QTableWidget(0, 8, central);
    table_->setHorizontalHeaderLabels({"ID", "Производитель", "Модель", "Локация", "Серийный", "Инвентарный", "Статус", "Дата установки"});
    table_->setEditTriggers(QAbstractItemView::NoEditTriggers);
    table_->setSelectionBehavior(QAbstractItemView::SelectRows);
    table_->setSelectionMode(QAbstractItemView::SingleSelection);
    table_->setSortingEnabled(true);
    table_->setColumnHidden(0, true);
    table_->setAlternatingRowColors(true);
    table_->setShowGrid(false);
    table_->verticalHeader()->setVisible(false);
    table_->horizontalHeader()->setStretchLastSection(true);
    // Больше места для «Производитель», чтобы имена не обрезались.
    table_->setColumnWidth(1, 240);
    layout->addWidget(table_);

    setCentralWidget(central);
    resize(900, 600);

    connect(refreshAction_, &QAction::triggered, this, &MainWindow::loadDevices);
    connect(searchEdit_, &QLineEdit::returnPressed, this, &MainWindow::loadDevices);
    connect(logoutAction_, &QAction::triggered, this, &MainWindow::logoutRequested);

    if (themeAction_) {
        connect(themeAction_, &QAction::toggled, this, [](bool enabled) {
            ThemeManager::setTheme(enabled ? ThemeManager::Theme::Dark : ThemeManager::Theme::Light);
        });
    }

    if (vendorsAction_) {
        connect(vendorsAction_, &QAction::triggered, this, [this]() {
            VendorsDialog dlg(apiClient_, this);
            dlg.exec();
        });
    }
    if (modelsAction_) {
        connect(modelsAction_, &QAction::triggered, this, [this]() {
            ModelsDialog dlg(apiClient_, this);
            dlg.exec();
        });
    }
    if (locationsAction_) {
        connect(locationsAction_, &QAction::triggered, this, [this]() {
            LocationsDialog dlg(apiClient_, this);
            dlg.exec();
        });
    }

    if (usersAction_) {
        connect(usersAction_, &QAction::triggered, this, [this]() {
            UsersDialog dlg(apiClient_, this);
            dlg.exec();
        });
    }

    connect(addAction_, &QAction::triggered, this, [this]() {
        if (!apiClient_) {
            UiUtils::critical(this, "Ошибка", "API клиент не инициализирован");
            return;
        }

        DeviceDialog dlg(apiClient_, this);
        QString err;
        if (!dlg.loadReferenceData(err)) {
            UiUtils::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить справочники" : err);
            return;
        }

        if (dlg.exec() != QDialog::Accepted) {
            return;
        }

        const qint64 modelId = dlg.selectedModelId();
        if (modelId <= 0) {
            UiUtils::information(this, "Добавление", "Выберите модель");
            return;
        }

        if (!apiClient_->createDevice(modelId, dlg.selectedLocationId(), dlg.serialNumber(), dlg.inventoryNumber(),
                                      dlg.status(), dlg.installedAt(), dlg.description(), err)) {
            UiUtils::warning(this, "Не удалось добавить", err.isEmpty() ? "Ошибка" : err);
            return;
        }

        loadDevices();
    });

    connect(editAction_, &QAction::triggered, this, [this]() {
        const qint64 id = selectedDeviceId();
        if (id <= 0) {
            UiUtils::information(this, "Редактирование", "Выберите устройство в таблице");
            return;
        }
        if (!apiClient_) {
            UiUtils::critical(this, "Ошибка", "API клиент не инициализирован");
            return;
        }

        QString err;
        DeviceDetails details;
        if (!apiClient_->getDevice(id, details, err)) {
            UiUtils::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить устройство" : err);
            return;
        }

        DeviceDialog dlg(apiClient_, this);
        if (!dlg.loadReferenceData(err)) {
            UiUtils::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить справочники" : err);
            return;
        }
        dlg.setInitialDevice(details);

        if (dlg.exec() != QDialog::Accepted) {
            return;
        }

        const qint64 modelId = dlg.selectedModelId();
        if (modelId <= 0) {
            UiUtils::information(this, "Редактирование", "Выберите модель");
            return;
        }

        if (!apiClient_->updateDevice(id, modelId, dlg.selectedLocationId(), dlg.serialNumber(), dlg.inventoryNumber(),
                                      dlg.status(), dlg.installedAt(), dlg.description(), err)) {
            UiUtils::warning(this, "Не удалось сохранить", err.isEmpty() ? "Ошибка" : err);
            return;
        }

        loadDevices();
    });

    connect(deleteAction_, &QAction::triggered, this, [this]() {
        const qint64 id = selectedDeviceId();
        if (id <= 0) {
            UiUtils::information(this, "Удаление", "Выберите устройство в таблице");
            return;
        }

        const auto answer = UiUtils::question(this, "Удаление", "Удалить выбранное устройство?");
        if (answer != QMessageBox::Yes) {
            return;
        }

        if (!apiClient_) {
            UiUtils::critical(this, "Ошибка", "API клиент не инициализирован");
            return;
        }

        QString err;
        if (!apiClient_->deleteDevice(id, err)) {
            UiUtils::warning(this, "Не удалось удалить", err.isEmpty() ? "Ошибка" : err);
            return;
        }
        loadDevices();
    });
}

void MainWindow::loadDevices() {
    if (!apiClient_) {
        return;
    }

    QList<DeviceItem> devices;
    QString err;
    if (!apiClient_->listDevices(searchEdit_ ? searchEdit_->text().trimmed() : QString(), devices, err)) {
        UiUtils::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить устройства" : err);
        return;
    }

    devices_ = devices;

    const bool wasSorting = table_->isSortingEnabled();
    table_->setSortingEnabled(false);

    table_->setRowCount(0);
    table_->setRowCount(devices.size());

    for (int i = 0; i < devices.size(); ++i) {
        const auto& d = devices[i];
        table_->setItem(i, 0, new QTableWidgetItem(QString::number(d.id)));
        table_->setItem(i, 1, new QTableWidgetItem(d.vendorName));
        table_->setItem(i, 2, new QTableWidgetItem(d.modelName));
        table_->setItem(i, 3, new QTableWidgetItem(d.locationName));
        table_->setItem(i, 4, new QTableWidgetItem(d.serialNumber));
        table_->setItem(i, 5, new QTableWidgetItem(d.inventoryNumber));
        table_->setItem(i, 6, new QTableWidgetItem(d.status));
        table_->setItem(i, 7, new QTableWidgetItem(d.installedAt));
    }

    table_->setSortingEnabled(wasSorting);
}

qint64 MainWindow::selectedDeviceId() const {
    if (!table_) {
        return 0;
    }
    const auto ranges = table_->selectedRanges();
    if (ranges.isEmpty()) {
        return 0;
    }
    const int row = ranges.first().topRow();
    const QTableWidgetItem* item = table_->item(row, 0);
    if (!item) {
        return 0;
    }
    return item->text().toLongLong();
}
