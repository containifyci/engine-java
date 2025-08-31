package com.example;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;

/**
 * Simple Hello World Servlet for integration testing
 */
public class HelloWorldServlet extends HttpServlet {

    private static final long serialVersionUID = 1L;
    private String message;

    @Override
    public void init() throws ServletException {
        message = "Hello World from Tomcat Servlet!";
    }

    @Override
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        
        response.setContentType("text/html;charset=UTF-8");
        
        try (PrintWriter out = response.getWriter()) {
            out.println("<!DOCTYPE html>");
            out.println("<html>");
            out.println("<head>");
            out.println("<title>Hello World Servlet</title>");
            out.println("</head>");
            out.println("<body>");
            out.println("<h1>" + message + "</h1>");
            out.println("<p>This servlet is working correctly!</p>");
            out.println("<p>Request URI: " + request.getRequestURI() + "</p>");
            out.println("<p>Servlet Path: " + request.getServletPath() + "</p>");
            out.println("</body>");
            out.println("</html>");
        }
    }

    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        doGet(request, response);
    }

    @Override
    public void destroy() {
        // Clean up resources if needed
    }
}