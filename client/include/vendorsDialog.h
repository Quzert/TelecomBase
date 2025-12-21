#pragma once

#include <QDialog>
#include <QList>

#include "models.h"

class ApiClient;
class QTableWidget;
class QAction;

class VendorsDialog final : public QDialog {
    Q_OBJECT

public:
    explicit VendorsDialog(ApiClient* apiClient, QWidget* parent = nullptr);

private:
    void buildUi();
    void reload();
    qint64 selectedId() const;

    void addVendor();
    void editVendor();
    void deleteVendor();

    ApiClient* apiClient_;
    QList<VendorItem> vendors_;

    QTableWidget* table_;
    QAction* addAction_;
    QAction* editAction_;
    QAction* deleteAction_;
    QAction* refreshAction_;
};
