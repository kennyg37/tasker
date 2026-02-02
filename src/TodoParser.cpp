#include "TodoParser.h"
#include <fstream>
#include <sstream>
#include <iostream>
#include <cstdlib>
#include <algorithm>
#include <cctype>

TodoParser::TodoParser() {
    filePath = getFilePathFromEnv();
}

std::string TodoParser::getFilePathFromEnv() const {
    const char* envPath = std::getenv("TODO_FILE_PATH");
    if (envPath == nullptr) {
        std::cerr << "Error: TODO_FILE_PATH environment variable not set!" << std::endl;
        std::cerr << "Please set it to the path of your todo.txt file." << std::endl;
        return "";
    }
    return std::string(envPath);
}

bool TodoParser::isDateLine(const std::string& line) const {
    // Simple check: line contains "/" and is relatively short (date format: DD/MM/YYYY)
    if (line.empty()) return false;
    
    // Trim whitespace
    std::string trimmed = line;
    trimmed.erase(0, trimmed.find_first_not_of(" \t\r\n"));
    trimmed.erase(trimmed.find_last_not_of(" \t\r\n") + 1);
    
    // Check if it looks like a date (contains / and digits)
    return (trimmed.find('/') != std::string::npos && 
            trimmed.length() >= 8 && 
            trimmed.length() <= 12);
}

TaskStatus TodoParser::parseStatus(const std::string& statusStr) const {
    std::string lower = statusStr;
    std::transform(lower.begin(), lower.end(), lower.begin(), ::tolower);
    
    if (lower.find("done") != std::string::npos) {
        return TaskStatus::Done;
    } else if (lower.find("postponed") != std::string::npos) {
        return TaskStatus::Postponed;
    } else if (lower.find("pending") != std::string::npos) {
        return TaskStatus::Pending;
    } else {
        return TaskStatus::Other;
    }
}

void TodoParser::parseLine(const std::string& line, const std::string& currentDate, std::vector<Task>& tasks) const {
    if (line.empty()) return;
    
    // Format: "1. Description - Status"
    size_t dotPos = line.find('.');
    if (dotPos == std::string::npos) return;
    
    // Extract task number
    std::string numberStr = line.substr(0, dotPos);
    numberStr.erase(0, numberStr.find_first_not_of(" \t"));
    int taskNumber = 0;
    try {
        taskNumber = std::stoi(numberStr);
    } catch (...) {
        return; // Not a valid task line
    }
    
    // Find the status separator
    size_t dashPos = line.find('-', dotPos);
    if (dashPos == std::string::npos) {
        // No status, assume pending
        std::string description = line.substr(dotPos + 1);
        description.erase(0, description.find_first_not_of(" \t"));
        description.erase(description.find_last_not_of(" \t\r\n") + 1);
        tasks.emplace_back(currentDate, taskNumber, description, TaskStatus::Pending);
        return;
    }
    
    // Extract description and status
    std::string description = line.substr(dotPos + 1, dashPos - dotPos - 1);
    description.erase(0, description.find_first_not_of(" \t"));
    description.erase(description.find_last_not_of(" \t") + 1);
    
    std::string statusStr = line.substr(dashPos + 1);
    statusStr.erase(0, statusStr.find_first_not_of(" \t"));
    statusStr.erase(statusStr.find_last_not_of(" \t\r\n") + 1);
    
    TaskStatus status = parseStatus(statusStr);
    tasks.emplace_back(currentDate, taskNumber, description, status);
}

std::vector<Task> TodoParser::parse() {
    std::vector<Task> tasks;
    
    if (filePath.empty()) {
        return tasks;
    }
    
    std::ifstream file(filePath);
    if (!file.is_open()) {
        std::cerr << "Error: Could not open file: " << filePath << std::endl;
        return tasks;
    }
    
    std::string line;
    std::string currentDate;
    
    while (std::getline(file, line)) {
        // Remove carriage return if present (Windows line endings)
        if (!line.empty() && line.back() == '\r') {
            line.pop_back();
        }
        
        if (isDateLine(line)) {
            // Trim whitespace
            currentDate = line;
            currentDate.erase(0, currentDate.find_first_not_of(" \t\r\n"));
            currentDate.erase(currentDate.find_last_not_of(" \t\r\n") + 1);
        } else if (!currentDate.empty()) {
            parseLine(line, currentDate, tasks);
        }
    }
    
    file.close();
    return tasks;
}
