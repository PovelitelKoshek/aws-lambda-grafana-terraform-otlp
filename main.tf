terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

resource "aws_iam_role" "lambda_role" {
  name = "lambda-grafana-terraform-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "lambda_function" {
  function_name = "lambda-grafana-terraform-direct"
  role          = aws_iam_role.lambda_role.arn

  filename         = "${path.module}/lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/lambda.zip")

  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["x86_64"]

  memory_size = 128
  timeout     = 30

  environment {
    variables = {
      OTEL_SERVICE_NAME           = "lambda-grafana-terraform"
      OTEL_EXPORTER_OTLP_PROTOCOL = var.otel_exporter_otlp_protocol
      OTEL_EXPORTER_OTLP_ENDPOINT = var.otel_exporter_otlp_endpoint
      OTEL_EXPORTER_OTLP_HEADERS  = var.otel_exporter_otlp_headers
      OTEL_RESOURCE_ATTRIBUTES    = "deployment.environment=student-demo,service.namespace=aws-lambda-grafana,service.version=1.0.0"
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution
  ]
}
