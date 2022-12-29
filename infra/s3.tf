resource "aws_s3_bucket" "random" {
  bucket = "ew-pad-bucket"
  tags = {
    Name        = "pad bucket"
  }
}

resource "aws_s3_bucket_acl" "random" {
  bucket = aws_s3_bucket.random.id
  acl    = "private"
}

//make a website bucket that's just SSG 
