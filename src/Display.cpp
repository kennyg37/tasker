#include "Display.h"
#include <iostream>
#include <iomanip>

const std::string Display::COLOR_RESET = "\033[0m";
const std::string Display::COLOR_GREEN = "\033[32m";
const std::string Display::COLOR_YELLOW = "\033[33m";
const std::string Display::COLOR_RED = "\033[31m";
const std::string Display::COLOR_BLUE = "\033[34m";
const std::string Display::COLOR_CYAN = "\033[36m";
const std::string Display::COLOR_BOLD = "\033[1m";

std::string Display::getColorForStatus(TaskStatus status) const {
    switch (status) {
        case TaskStatus::Done:
            return COLOR_GREEN;
        case TaskStatus::Postponed:
            return COLOR_YELLOW;
        case TaskStatus::Pending:
            return COLOR_BLUE;
        case TaskStatus::Other:
            return COLOR_RED;
        default:
            return COLOR_RESET;
    }
}

void Display::printHeader() const {
    std::cout << COLOR_BOLD << COLOR_CYAN;
    std::cout << "+--------------------------------------------------------------------------+\n";
    std::cout << COLOR_RESET;
}

void Display::printSeparator() const {
    std::cout << COLOR_CYAN;
    std::cout << "+--------------------------------------------------------------------------+\n";
    std::cout << COLOR_RESET;
}

void Display::showTasks(const std::vector<Task>& tasks, const std::string& title) const {
    printHeader();
    std::cout << COLOR_BOLD << "|  " << std::setw(70) << std::left << title << "  |\n" << COLOR_RESET;
    printSeparator();
    
    if (tasks.empty()) {
        std::cout << "|  " << std::setw(70) << std::left << "No tasks found." << "  |\n";
    } else {
        for (const auto& task : tasks) {
            std::string statusColor = getColorForStatus(task.getStatus());
            std::ostringstream line;
            line << std::setw(12) << task.getDate() << " | "
                 << task.getTaskNumber() << ". "
                 << std::setw(35) << std::left << task.getDescription()
                 << " [" << task.getStatusString() << "]";
            
            std::cout << "|  " << statusColor << std::setw(70) << std::left << line.str() 
                     << COLOR_RESET << "  |\n";
        }
    }
    
    std::cout << COLOR_CYAN;
    std::cout << "+--------------------------------------------------------------------------+\n";
    std::cout << COLOR_RESET;
}

void Display::showStatistics(const TodoManager& manager) const {
    printHeader();
    std::cout << COLOR_BOLD << "|  " << std::setw(70) << std::left << "Statistics" << "  |\n" << COLOR_RESET;
    printSeparator();
    
    int total = manager.getTotalCount();
    int done = manager.getCountByStatus(TaskStatus::Done);
    int postponed = manager.getCountByStatus(TaskStatus::Postponed);
    int pending = manager.getCountByStatus(TaskStatus::Pending);
    int other = manager.getCountByStatus(TaskStatus::Other);
    
    std::cout << "|  " << std::setw(70) << std::left << ("Total Tasks: " + std::to_string(total)) << "  |\n";
    std::cout << "|  " << COLOR_GREEN << std::setw(70) << std::left << ("Done: " + std::to_string(done)) 
             << COLOR_RESET << "  |\n";
    std::cout << "|  " << COLOR_YELLOW << std::setw(70) << std::left << ("Postponed: " + std::to_string(postponed)) 
             << COLOR_RESET << "  |\n";
    std::cout << "|  " << COLOR_BLUE << std::setw(70) << std::left << ("Pending: " + std::to_string(pending)) 
             << COLOR_RESET << "  |\n";
    std::cout << "|  " << std::setw(70) << std::left << ("Other: " + std::to_string(other)) << "  |\n";
    
    printSeparator();
    
    // Completion percentage
    double completionRate = total > 0 ? (static_cast<double>(done) / total) * 100.0 : 0.0;
    std::ostringstream percentLine;
    percentLine << "Completion Rate: " << std::fixed << std::setprecision(1) << completionRate << "%";
    std::cout << "|  " << COLOR_BOLD << std::setw(70) << std::left << percentLine.str() 
             << COLOR_RESET << "  |\n";
    
    // Tasks by date
    printSeparator();
    std::cout << "|  " << COLOR_BOLD << std::setw(70) << std::left << "Tasks by Date:" << COLOR_RESET << "  |\n";
    
    auto dateCount = manager.getTaskCountByDate();
    for (const auto& [date, count] : dateCount) {
        std::ostringstream dateLine;
        dateLine << "  " << date << ": " << count << " task(s)";
        std::cout << "|  " << std::setw(70) << std::left << dateLine.str() << "  |\n";
    }
    
    std::cout << COLOR_CYAN;
    std::cout << "+--------------------------------------------------------------------------+\n";
    std::cout << COLOR_RESET;
}

void Display::showHelp() const {
    printHeader();
    std::cout << COLOR_BOLD << "|  " << std::setw(70) << std::left << "Tasker CLI - Help" << "  |\n" << COLOR_RESET;
    printSeparator();
    
    std::cout << "|  " << COLOR_BOLD << "Usage:" << COLOR_RESET << std::setw(62) << " " << "  |\n";
    std::cout << "|    tasker <command> [options]" << std::setw(42) << " " << "  |\n";
    std::cout << "|" << std::setw(76) << " " << "|\n";
    std::cout << "|  " << COLOR_BOLD << "Commands:" << COLOR_RESET << std::setw(60) << " " << "  |\n";
    std::cout << "|    " << COLOR_CYAN << "list" << COLOR_RESET << "               Show all tasks" << std::setw(34) << " " << "  |\n";
    std::cout << "|    " << COLOR_CYAN << "stats" << COLOR_RESET << "              Show statistics" << std::setw(34) << " " << "  |\n";
    std::cout << "|    " << COLOR_CYAN << "filter date <DD/MM/YYYY>" << COLOR_RESET << "  Filter tasks by date" << std::setw(19) << " " << "  |\n";
    std::cout << "|    " << COLOR_CYAN << "filter status <STATUS>" << COLOR_RESET << "    Filter by status (done/postponed/pending)" << std::setw(1) << " " << "  |\n";
    std::cout << "|    " << COLOR_CYAN << "search <keyword>" << COLOR_RESET << "      Search tasks by keyword" << std::setw(27) << " " << "  |\n";
    std::cout << "|    " << COLOR_CYAN << "help" << COLOR_RESET << "               Show this help message" << std::setw(28) << " " << "  |\n";
    std::cout << "|" << std::setw(76) << " " << "|\n";
    std::cout << "|  " << COLOR_BOLD << "Examples:" << COLOR_RESET << std::setw(58) << " " << "  |\n";
    std::cout << "|    tasker list" << std::setw(58) << " " << "  |\n";
    std::cout << "|    tasker filter date 23/1/2025" << std::setw(41) << " " << "  |\n";
    std::cout << "|    tasker filter status done" << std::setw(44) << " " << "  |\n";
    std::cout << "|    tasker search polymorphism" << std::setw(43) << " " << "  |\n";
    std::cout << "|    tasker stats" << std::setw(57) << " " << "  |\n";
    
    std::cout << COLOR_CYAN;
    std::cout << "+--------------------------------------------------------------------------+\n";
    std::cout << COLOR_RESET;
}
